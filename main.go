package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	universalDeserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
)

func main() {
	certFile := "/etc/webhook/certs/tls.crt"
	keyFile := "/etc/webhook/certs/tls.key"

	router := mux.NewRouter()
	router.Path("/mutate").Handler(http.HandlerFunc(MutatingHandler))
	router.Path("/validate").Handler(http.HandlerFunc(ValidatingHandler))

	log.Fatal(http.ListenAndServeTLS(":"+strconv.Itoa(8443), certFile, keyFile, router))
}

func getAdmissionReviewReq(r *http.Request) (*v1beta1.AdmissionReview, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var admissionReviewReq v1beta1.AdmissionReview
	if _, _, err := universalDeserializer.Decode(body, nil, &admissionReviewReq); err != nil {
		return nil, err
	} else if admissionReviewReq.Request == nil {
		return nil, errors.New("malformed admission review: request is nil: " + err.Error())
	}

	admissionReviewReq.Response = &v1beta1.AdmissionResponse{
		UID: admissionReviewReq.Request.UID,
		// Allowed: true,
	}

	log.Printf("Type: %v  Event: %v Name: %v ",
		admissionReviewReq.Request.Kind,
		admissionReviewReq.Request.Operation,
		admissionReviewReq.Request.Name,
	)
	return &admissionReviewReq, nil
}

func MutatingHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Mutating request")
	admissionReviewReq, err := getAdmissionReviewReq(r)
	if err != nil {
		http.Error(w, "malformed admission review: request is nil: "+err.Error(), http.StatusBadRequest)
	}
	op := admissionReviewReq.Request.Operation
	if op == "CREATE" {
		var patch string
		patchType := v1.PatchTypeJSONPatch
		patch = `[{"op":"add","path":"/metadata/labels","value":{"gophers":"bangalore"}}]`

		admissionReviewReq.Response.Allowed = true
		admissionReviewReq.Response.PatchType = (*v1beta1.PatchType)(&patchType)
		admissionReviewReq.Response.Patch = []byte(patch)
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(admissionReviewReq)
	if err != nil {
		fmt.Errorf(err.Error())
	}
}

func ValidatingHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("validating request")
	admissionReviewReq, err := getAdmissionReviewReq(r)
	if err != nil {
		http.Error(w, "malformed admission review: request is nil: "+err.Error(), http.StatusBadRequest)
	}

	if admissionReviewReq.Request.Kind.Kind != "Workflow" {
		errMessage := "expected Workflow got " + admissionReviewReq.Request.Kind.Kind
		log.Fatalf(errMessage)
		http.Error(w, errMessage, http.StatusBadRequest)
		return
	}

	op := admissionReviewReq.Request.Operation
	if op == "CREATE" || op == "UPDATE" {
		if !strings.HasPrefix(admissionReviewReq.Request.Name, "gophers") {
			admissionReviewReq.Response.Allowed = false
			admissionReviewReq.Response.Result = &metav1.Status{
				Message: "Bro, you are in gophers meetup!",
			}
			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(admissionReviewReq)
			if err != nil {
				fmt.Errorf(err.Error())
			}
			return
		}

		// persist to ES
		// var wf *wfv1.Workflow
		// err = json.Unmarshal(admissionReviewReq.Request.Object.Raw, &wf)
		// if err != nil {
		// 	log.Fatalf(fmt.Sprintf("could not unmarshal WorkflowTemplate on admission request: " + err.Error()))
		// 	return
		// }

		// _, err = esClient.Index().Index("workflows").Id(wf.ObjectMeta.Name).BodyJson(wf).Do(context.Background())
		// if err != nil {
		// 	log.Fatalf("failed to upsert to database: " + err.Error())
		// 	return
		// }
	}

	admissionReviewReq.Response.Allowed = true
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(admissionReviewReq)
	if err != nil {
		fmt.Errorf(err.Error())
	}
}
