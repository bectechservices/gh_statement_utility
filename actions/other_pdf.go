package actions

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"ng-statement-app/models"
	"path/filepath"

	"github.com/gobuffalo/buffalo/binding"
	"github.com/gobuffalo/envy"
)

func MakeRestCallToStampifyAPI(user *models.User, fileName string, file *binding.File, responseChan chan []byte, errChan chan error) {
	defer file.Close()
	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer

	writer := multipart.NewWriter(&requestBody)
	// Create a form file field for the file
	fileFormField, err := writer.CreateFormFile("file", filepath.Base(file.Filename))
	if err != nil {
		responseChan <- nil
		errChan <- errors.New("failed to create form file")
		return
	}

	// Copy the file content to the form file field
	_, err = io.Copy(fileFormField, file)
	if err != nil {
		log.Println(err)
		responseChan <- nil
		errChan <- errors.New("failed to copy file content")
		return
	}

	writer.WriteField("file_name", fileName)
	writer.WriteField("file_type", "PDF")
	writer.WriteField("user_id", user.ID.String())

	err = writer.Close()
	if err != nil {
		responseChan <- nil
		errChan <- errors.New("failed to close writer")
		return
	}

	// Create a new request using http with multipart form data
	stampifyPdfAPIHost := envy.Get("STAMPIFYPDF_API_HOST", "")
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/generate-stamps", stampifyPdfAPIHost), &requestBody)
	if err != nil {
		log.Println(err)
		responseChan <- nil
		errChan <- errors.New("failed to create request")
		return
	}
	log.Println("Request: ", req.URL)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	// Make the request
	client := &http.Client{}
	log.Println("making rest call to stampifypdf api....")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		responseChan <- nil
		errChan <- errors.New("failed to generate stamp on file")
		return
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		log.Println(bodyString)
		responseChan <- nil
		errChan <- errors.New("failed to generate stamp on file")
		return
	}

	responseContents, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	if err != nil || responseContents == nil {
		log.Println(err)
		responseChan <- nil
		errChan <- errors.New("failed to generate stamp on file")
		return
	}

	responseChan <- responseContents
	errChan <- nil
}

func MakeOtherStampifyPdfAPI(user *models.User, files []binding.File) {
	// otherusers := models.UserStampDetail{}
	var requestBody bytes.Buffer
	for _, file := range files {
		fmt.Println("################# File PDF OTHERS :", file.Filename)
		if !file.Valid() {
			continue
		}
		writer := multipart.NewWriter(&requestBody)
		fmt.Println("################# writer!!!!!!!!!! :", writer)
		// Create a form file field for the file\
		fmt.Println("################# File PATH >>>>>>>> :", filepath.Base(file.Filename))
		fileFormField, err := writer.CreateFormFile("file", filepath.Base(file.Filename))
		fmt.Println("################# fileFormField <<<<<>>>>> :", fileFormField)
		if err != nil {
			log.Println(err)
		}
		// Copy the file content to the form file field
		_, err = io.Copy(fileFormField, file)
		if err != nil {
			log.Println(err)
		}
		fmt.Println("##########file_name & file_type & user_id", file.Filename, user.ID.String())
		writer.WriteField("file_name", file.Filename)
		writer.WriteField("file_type", "PDF")
		writer.WriteField("user_id", user.ID.String())

		err = writer.Close()
		if err != nil {
			log.Println(err)
		}

		// Create a new request using http with multipart form data
		stampifyPdfAPIHost := envy.Get("STAMPIFYPDF_API_HOST", "")
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/generate-stamps", stampifyPdfAPIHost), &requestBody)
		if err != nil {
			log.Println(err)
		}
		log.Println("Request: ", req.URL)
		client := &http.Client{}
		log.Println("making rest call to stampifypdf api....")
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
		}
		defer resp.Body.Close()

		// Check if the response status is OK
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyString := string(bodyBytes)

			fmt.Println("##########<<<<<<< bodyBytes >>>>>>>", bodyBytes)

			log.Println(bodyString)
			log.Println(err)
		}

		responseContents, err := io.ReadAll(resp.Body)
		fmt.Println("##########<<<<<<< responseContents >>>>>>>", responseContents)
		if err != nil {
			fmt.Println("Error reading response body:", err)

		}

		if err != nil || responseContents == nil {
			log.Println(err)

		}

	}

}
