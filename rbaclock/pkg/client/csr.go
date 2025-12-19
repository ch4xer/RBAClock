package client

import (
	"context"

	certv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListCertificateSigningRequests() []certv1.CertificateSigningRequest {
	csrs, _ := Client().CertificatesV1().CertificateSigningRequests().List(context.TODO(), metav1.ListOptions{})

	return csrs.Items
}
