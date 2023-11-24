package main

import (
	"encoding/json"
	"io"
	v1 "k8s.io/api/admission/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("HandleRoot!"))
}

func HandleUploadCertToWaf(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal("error while reading admission request body", err)
	}

	admissionReviewReq := deserializeToAdmissionReview(w, body)

	log.Println("received admission review")
	log.Println(string(admissionReviewReq.Request.Object.Raw))

	var secret apiv1.Secret

	err = json.Unmarshal(admissionReviewReq.Request.Object.Raw, &secret)

	certId := CreateOrUpdateCertificate(secret)

	patchBytes := createIdPatch(secret, certId)

	responseBytes := createAdmissionResponse(admissionReviewReq, patchBytes, err)

	w.Write(responseBytes)
}

func deserializeToAdmissionReview(w http.ResponseWriter, body []byte) v1.AdmissionReview {
	var admissionReviewReq v1.AdmissionReview

	if _, _, err := universalDeserializer.Decode(body, nil, &admissionReviewReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Fatal("could not deserialize admission request", err)
	} else if admissionReviewReq.Request == nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Fatal("malformed admission review: request is nil")
	}

	log.Printf("Type: %v \t Event: %v \t Name: %v \t RequestKind: %v \t UID: %v \t UserInfo: %v \n",
		admissionReviewReq.Request.Kind,
		admissionReviewReq.Request.Operation,
		admissionReviewReq.Request.Name,
		admissionReviewReq.Request.RequestKind,
		admissionReviewReq.Request.UID,
		admissionReviewReq.Request.UserInfo,
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
		log.Fatal("marshaling admission response failed", err)
	}

	log.Println(string(bytes))
	return bytes
}

func applyPatchesToAdmissionResponse(admissionReviewResponse v1.AdmissionReview, patchBytes []byte) {
	admissionReviewResponse.Response.Patch = patchBytes
	pt := v1.PatchTypeJSONPatch
	admissionReviewResponse.Response.PatchType = &pt
}

func createIdPatch(secret apiv1.Secret, id string) []byte {
	var patches []patchOperation

	labels := secret.ObjectMeta.Labels
	labels["cert-waf-id"] = id

	patches = append(patches, patchOperation{
		Op:    "add",
		Path:  "/metadata/labels",
		Value: labels,
	})

	patchBytes, err := json.Marshal(patches)

	if err != nil {
		log.Fatal("could not marshal JSON patch", err)
	}
	return patchBytes
}
