
Start a docker container
```
docker run -it --rm -v ${PWD}:/work -w /work debian bash
```


Download cloudflare ssl tools
```
apt-get update && apt-get install -y curl &&
curl -L https://github.com/cloudflare/cfssl/releases/download/v1.5.0/cfssl_1.5.0_linux_amd64 -o /usr/local/bin/cfssl && \
curl -L https://github.com/cloudflare/cfssl/releases/download/v1.5.0/cfssljson_1.5.0_linux_amd64 -o /usr/local/bin/cfssljson && \
chmod +x /usr/local/bin/cfssl && \
chmod +x /usr/local/bin/cfssljson
```

generate ca in /tmp
```
cfssl gencert -initca ./ca-csr.json | cfssljson -bare /tmp/ca
```

generate certificate in /tmp
```
 cfssl gencert \
  -ca=/tmp/ca.pem \
  -ca-key=/tmp/ca-key.pem \
  -config=./ca-config.json \
  -hostname="jobsync,jobsync.default.svc.cluster.local,jobsync.default.svc,localhost,127.0.0.1" \
  -profile=default \
  ./ca-csr.json | cfssljson -bare /tmp/jobsync 
```

Create the secret file
```
cat <<EOF > ./jobsync-tls.yaml
apiVersion: v1
kind: Secret
metadata:
  name: jobsync-tls
type: Opaque
data:
  tls.crt: $(cat /tmp/jobsync.pem | base64 | tr -d '\n')
  tls.key: $(cat /tmp/jobsync-key.pem | base64 | tr -d '\n') 
EOF
```


#generate CA Bundle + inject into template
```
ca_pem_b64="$(openssl base64 -A <"/tmp/ca.pem")"
```

with this the k8s will know how to verify the bundle since it will have the ca certificate
```
sed -e 's@${CA_PEM_B64}@'"$ca_pem_b64"'@g' <"webhook-template.yaml" \
    > webhook.yaml
```