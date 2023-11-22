package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	v1 "k8s.io/api/admission/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("HandleRoot!"))
}

func HandleMutate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	err = os.WriteFile("/tmp/request", body, 0644)
	if err != nil {
		panic(err.Error())
	}

	admissionReviewReq := deserializeToAdmissionReview(w, body)

	var secret apiv1.Secret

	err = json.Unmarshal(admissionReviewReq.Request.Object.Raw, &secret)

	if err != nil {
		fmt.Errorf("could not unmarshal pod on admission request: %v", err)
	}

	fmt.Println("certificate tls.crt: " + string(secret.Data["tls.crt"]))
	fmt.Println("certificate tls.key: " + string(secret.Data["tls.key"]))

	patchBytes := createPatches(secret)

	responseBytes := createAdmissionResponse(admissionReviewReq, patchBytes, err)

	w.Write(responseBytes)
}

func HandleUploadCertToWaf(w http.ResponseWriter, r *http.Request) {

}

func deserializeToAdmissionReview(w http.ResponseWriter, body []byte) v1.AdmissionReview {
	var admissionReviewReq v1.AdmissionReview

	if _, _, err := universalDeserializer.Decode(body, nil, &admissionReviewReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Errorf("could not deserialize request: %v", err)
	} else if admissionReviewReq.Request == nil {
		w.WriteHeader(http.StatusBadRequest)
		errors.New("malformed admission review: request is nil")
	}

	fmt.Printf("Type: %v \t Event: %v \t Name: %v \n",
		admissionReviewReq.Request.Kind,
		admissionReviewReq.Request.Operation,
		admissionReviewReq.Request.Name,
	)
	return admissionReviewReq
}

func createAdmissionResponse(admissionReviewReq v1.AdmissionReview, patchBytes []byte, err error) []byte {
	admissionReviewResponse := v1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &v1.AdmissionResponse{
			UID:     admissionReviewReq.Request.UID,
			Allowed: true,
		},
	}

	applyPatchesToAdmissionResponse(admissionReviewResponse, patchBytes)

	bytes, err := json.MarshalIndent(&admissionReviewResponse, "", "    ")
	if err != nil {
		fmt.Errorf("marshaling response: %v", err)
	}

	fmt.Println(string(bytes))
	return bytes
}

func applyPatchesToAdmissionResponse(admissionReviewResponse v1.AdmissionReview, patchBytes []byte) {
	admissionReviewResponse.Response.Patch = patchBytes
	pt := v1.PatchTypeJSONPatch
	admissionReviewResponse.Response.PatchType = &pt
}

func createPatches(secret apiv1.Secret) []byte {
	var patches []patchOperation

	labels := secret.ObjectMeta.Labels
	labels["example-webhook"] = "it-worked"

	patches = append(patches, patchOperation{
		Op:    "add",
		Path:  "/metadata/labels",
		Value: labels,
	})

	patchBytes, err := json.Marshal(patches)

	if err != nil {
		fmt.Errorf("could not marshal JSON patch: %v", err)
	}
	return patchBytes
}
