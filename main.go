package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"time"

	"gocv.io/x/gocv"
)

var (
	everaiHost = flag.String("host", "", "EverAI host")
	refImg     = flag.String("ref", "", "Reference image")
)

func main() {
	flag.Parse()

	tmpDir, capturedImgPath := captureImage()
	defer os.Remove(tmpDir)

	verifyImage(capturedImgPath)
}

func captureImage() (string, string) {
	webcam, err := gocv.OpenVideoCapture(0)
	if err != nil {
		log.Fatalf("failed to open webcam: %+v", err)
	}
	defer webcam.Close()

	img := gocv.NewMat()
	defer img.Close()

	if ok := webcam.Read(&img); !ok {
		log.Fatal("failed to read from webcam")
	}

	tmpDir, err := ioutil.TempDir("", "xxx")
	if err != nil {
		log.Fatalf("failed to create temp directory: %+v", err)
	}

	imgPath := filepath.Join(tmpDir, "image.jpg")
	if ok := gocv.IMWrite(imgPath, img); !ok {
		log.Fatalf("failed to write webcam image to disk")
	}

	return tmpDir, imgPath
}

func verifyImage(capturedImgPath string) {
	req := buildRequest(capturedImgPath)

	startTime := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("failed to make request: %+v", err)
	}
	defer resp.Body.Close()

	totalTime := time.Since(startTime)
	log.Printf("Request took %s", totalTime)

	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		log.Fatalf("failed to copy response to stdout: %+v", err)
	}
}

func buildRequest(capturedImgPath string) *http.Request {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	if err := addMIMEPart(w, "image1", *refImg); err != nil {
		log.Fatal(err)
	}

	if err := addMIMEPart(w, "image2", capturedImgPath); err != nil {
		log.Fatal(err)
	}

	w.Close()

	apiURL := fmt.Sprintf("http://%s/api/verify", *everaiHost)
	req, err := http.NewRequest(http.MethodPost, apiURL, &buf)
	if err != nil {
		log.Fatalf("failed to create request: %+v", err)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func addMIMEPart(w *multipart.Writer, name, imgPath string) error {
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s.jpg"`, name, imgPath))
	header.Set("Content-Type", "image/jpeg")
	fw, err := w.CreatePart(header)

	if err != nil {
		return fmt.Errorf("failed to create MIME part writer: %+v", err)
	}

	img, err := os.Open(imgPath)
	if err != nil {
		return fmt.Errorf("failed to open image file %s: %+v", imgPath, err)
	}

	defer img.Close()
	n, err := io.Copy(fw, img)
	if err != nil {
		return fmt.Errorf("failed to add %s MIME part: %+v", name, err)
	}

	log.Printf("%s MIME part added (%d bytes)", name, n)
	return nil
}
