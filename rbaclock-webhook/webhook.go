package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/klog/v2"
	"slices"
)

const (
	newSchedulerName              = "rbaclock"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

var ignoredNamespaces = []string{
	metav1.NamespaceSystem,
}

type WebhookServer struct {
	server *http.Server
}

type WhSvrParameters struct {
	port     int    // webhook server port
	certFile string // path to the x509 certificate for https
	keyFile  string // path to the x509 private key matching `CertFile`
}

type patchOperation struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value,omitempty"`
}

func (whsvr *WebhookServer) serve(w http.ResponseWriter, r *http.Request) {
	var log bytes.Buffer
	// check request validity
	reqData, err := io.ReadAll(r.Body)
	if err != nil {
		log.WriteString("failed to read request")
		klog.Errorf("failed to read request: %v", err)
		http.Error(w, log.String(), http.StatusBadRequest)
		return
	}
	if len(reqData) == 0 {
		log.WriteString("empty request")
		klog.Errorf(log.String())
		http.Error(w, log.String(), http.StatusBadRequest)
		return
	}
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.WriteString(fmt.Sprintf("invalid Content-Type, expect `application/json`, got %s", contentType))
		klog.Errorf(log.String())
		http.Error(w, log.String(), http.StatusUnsupportedMediaType)
		return
	}
	// parse the admission request
	ar := admissionv1.AdmissionReview{}
	var admissionResponse *admissionv1.AdmissionResponse
	_, _, err = deserializer.Decode(reqData, nil, &ar)
	if err != nil {
		log.WriteString(fmt.Sprintf("\nCan't decode body,error info is : %s", err.Error()))
		klog.Errorf("Can't decode body: %v", err)
		admissionResponse = &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		fmt.Println(r.URL.Path)
		if r.URL.Path == "/mutate" {
			admissionResponse = mutate(&ar, &log)
		}
	}

	// webhook receives an AdmissionReview object containing an AdmissionRequest
	// it sends back an AdmissionReview object containing an AdmissionResponse.
	// the AdmissionResponse must include the UID of the coresponding AdmissionRequest.
	admissionReview := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
	}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		log.WriteString(fmt.Sprintf("\nCan't encode response,error info is : %s", err.Error()))
		klog.Errorf(log.String())
		http.Error(w, log.String(), http.StatusInternalServerError)
	} else {
		klog.Infof("Ready to write response ...")
		if _, err := w.Write(resp); err != nil {
			log.WriteString(fmt.Sprintf("\nCan't write response,error info is : %s", err.Error()))
			klog.Errorf(log.String())
			http.Error(w, log.String(), http.StatusInternalServerError)
		}
	}
	klog.Infof("Admission response sent.")
}

// func mutate(adminissionReview *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
//     return &admissionv1.AdmissionResponse{
//         Allowed: false,
//         Result: &metav1.Status{
//             Message: "You shall not pass!",
//         },
//     }
// }

func mutate(ar *admissionv1.AdmissionReview, log *bytes.Buffer) *admissionv1.AdmissionResponse {
	req := ar.Request
	var (
		oldSchedulerName                string
		objectMeta                      *metav1.ObjectMeta
		resourceNamespace, resourceName string
	)

	fmt.Fprintf(log, "\n======begin Admission for Namespace=[%v], Kind=[%v], Name=[%v]======", req.Namespace, req.Kind.Kind, req.Name)
	log.WriteString("\n>>>>>>" + req.Kind.Kind)

	switch req.Kind.Kind {
	case "Pod":
		var pod corev1.Pod
		if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
			fmt.Fprintf(log, "\nCould not unmarshal raw object: %v", err)
			klog.Errorf(log.String())
			return &admissionv1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		resourceNamespace, resourceName, objectMeta = pod.Namespace, pod.Name, &pod.ObjectMeta
		oldSchedulerName = pod.Spec.SchedulerName
	// case "Deployment":
	// 	var deployment appsv1.Deployment
	// 	if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
	// 		fmt.Fprintf(log, "\nCould not unmarshal raw object: %v", err)
	// 		klog.Errorf(log.String())
	// 		return &admissionv1.AdmissionResponse{
	// 			Result: &metav1.Status{
	// 				Message: err.Error(),
	// 			},
	// 		}
	// 	}
	// 	resourceName, resourceNamespace, objectMeta = deployment.Name, deployment.Namespace, &deployment.ObjectMeta
	// 	oldSchedulerName = deployment.Spec.Template.Spec.SchedulerName
	// 其他不支持的类型
	default:
		msg := fmt.Sprintf("\nNot support for this Kind of resource  %v", req.Kind.Kind)
		log.WriteString(msg)
		return &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: msg,
			},
		}
	}

	if !admissionRequired(ignoredNamespaces, oldSchedulerName, objectMeta) {
		fmt.Fprintf(log, "Skipping for %s/%s due to policy check", resourceNamespace, resourceName)
		klog.Infof(log.String())
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	patchBytes, err := createPatch(req.Kind.Kind, oldSchedulerName, newSchedulerName)
	if err != nil {
		fmt.Fprintf(log, "\nfailed to create patch: %v", err)
		klog.Errorf(log.String())
		return &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	fmt.Fprintf(log, "AdmissionResponse: patch=%v\n", string(patchBytes))
	klog.Infof(log.String())
	return &admissionv1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func admissionRequired(ignoredList []string, scheduler string, metadata *metav1.ObjectMeta) bool {
	// skip special kubernetes system namespaces
	if slices.Contains(ignoredList, metadata.Namespace) {
			klog.Infof("Skip %v in namespace:%v", metadata.Name, metadata.Namespace)
			return false
		}
	if scheduler == "rbaclock" {
		klog.Infof("Skip %v since it has rbaclock scheduler", metadata.Name)
		return false
	}
	return true
}

func createPatch(kind string, old string, new string) ([]byte, error) {
	var patch []patchOperation

	patch = append(patch, updateScheduler(kind, old, new)...)

	return json.Marshal(patch)
}

func updateScheduler(kind string, old string, new string) []patchOperation {
	var patch []patchOperation
	var path string
	if kind == "Pod" {
		path = "/spec/schedulerName"
	} else {
		path = "/spec/template/spec/schedulerName"
	}
	if old == "" {
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  path,
			Value: new,
		})
	} else {
		patch = append(patch, patchOperation{
			Op:    "replace",
			Path:  path,
			Value: new,
		})
	}
	return patch
}
