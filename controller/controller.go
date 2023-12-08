package controller

import (
	"encoding/json"
	"errors"
	"io"
	v1 "k8s.io/api/admission/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"log"
	"net/http"
	"waf-cert-uploader/service"
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

var createOrUpdateCertificate = func(secret apiv1.Secret) (*string, error) {
	return service.CreateOrUpdateCertificate(secret)
}

func HandleUploadCertToWaf(writer http.ResponseWriter, httpRequest *http.Request) {
	log.Println("received admission review")

	body, err := getRequestBody(httpRequest)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	admissionReview, err := deserializeToAdmissionReview(*body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	secret, err := getAdmissionReviewObject(*admissionReview)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	certId, wafServiceError := createOrUpdateCertificate(*secret)

	responseBytes, err := createResponseObject(wafServiceError, *admissionReview, *secret, certId)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	writeResponseObjectToConnection(writer, err, *responseBytes)
}

func writeResponseObjectToConnection(writer http.ResponseWriter, err error, responseBytes []byte) {
	_, err = writer.Write(responseBytes)
	if err != nil {
		log.Println("http reply failed", err)
	}
}

func createResponseObject(
	wafServiceError error,
	admissionReview v1.AdmissionReview,
	secret apiv1.Secret,
	certId *string) (*[]byte, error) {
	if wafServiceError != nil {
		log.Println("the admission review is rejected due to an error", wafServiceError)
		rejectResponse, err := createRejectAdmissionResponse(admissionReview)
		if err != nil {
			return nil, err
		}
		return rejectResponse, nil
	} else {
		patchBytes, err := createCertificateIdPatch(secret, *certId)
		if err != nil {
			return nil, err
		}
		patchedResponse, err := createPatchedAdmissionResponse(admissionReview, *patchBytes)
		if err != nil {
			return nil, err
		}
		return patchedResponse, nil
	}
}

func getRequestBody(httpRequest *http.Request) (*[]byte, error) {
	body, err := io.ReadAll(httpRequest.Body)
	if err != nil {
		log.Println("reading admission request body failed", err)
		return nil, err
	}
	return &body, nil
}

func getAdmissionReviewObject(admissionReview v1.AdmissionReview) (*apiv1.Secret, error) {
	var secret apiv1.Secret
	err := json.Unmarshal(admissionReview.Request.Object.Raw, &secret)
	if err != nil {
		log.Println("unmarshalling the admission review request object failed", err)
		return nil, err
	}
	return &secret, nil
}

func deserializeToAdmissionReview(body []byte) (*v1.AdmissionReview, error) {
	var admissionReview v1.AdmissionReview
	universalDeserializer := serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()

	if _, _, err := universalDeserializer.Decode(body, nil, &admissionReview); err != nil {
		log.Println("could not deserialize admission request", err)
		return nil, err
	} else if admissionReview.Request == nil {
		log.Println("malformed admission review: request is nil")
		return nil, errors.New("admission review was nil")
	}

	return &admissionReview, nil
}

func createPatchedAdmissionResponse(admissionReview v1.AdmissionReview, patchBytes []byte) (*[]byte, error) {
	admissionReviewResponse := createAdmissionReviewResponse(admissionReview, true)

	applyPatchesToAdmissionResponse(admissionReviewResponse, patchBytes)

	bytes, err := marshal(admissionReviewResponse)

	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func createRejectAdmissionResponse(admissionReview v1.AdmissionReview) (*[]byte, error) {
	admissionReviewResponse := createAdmissionReviewResponse(admissionReview, false)

	bytes, err := marshal(admissionReviewResponse)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func createAdmissionReviewResponse(admissionReview v1.AdmissionReview, allowed bool) v1.AdmissionReview {
	return v1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: admissionReview.APIVersion,
			Kind:       admissionReview.Kind,
		},
		Response: &v1.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: allowed,
		},
	}
}

func applyPatchesToAdmissionResponse(admissionReviewResponse v1.AdmissionReview, patchBytes []byte) {
	admissionReviewResponse.Response.Patch = patchBytes
	patchType := v1.PatchTypeJSONPatch
	admissionReviewResponse.Response.PatchType = &patchType
}

func createCertificateIdPatch(secret apiv1.Secret, id string) (*[]byte, error) {
	var patches []patchOperation

	annotations := secret.ObjectMeta.Labels
	annotations["cert-waf-id"] = id

	patches = append(patches, patchOperation{
		Op:    "add",
		Path:  "/metadata/annotations",
		Value: annotations,
	})

	patchBytes, err := marshal(patches)
	if err != nil {
		return nil, err
	}

	return patchBytes, nil
}

func marshal(any interface{}) (*[]byte, error) {
	bytes, err := json.Marshal(&any)
	if err != nil {
		log.Println("marshaling failed", err)
		return nil, err
	}
	return &bytes, nil
}
