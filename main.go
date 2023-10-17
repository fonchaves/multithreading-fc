package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

type ViaCEP struct {
	Cep        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	Uf         string `json:"uf"`
}

type ApiCEP struct {
	Cep        string `json:"code"`
	Logradouro string `json:"address"`
	Bairro     string `json:"district"`
	Localidade string `json:"city"`
	Uf         string `json:"state"`
}

type CepData struct {
	Cep        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	Uf         string `json:"uf"`
	Api        string `json:"api"`
}

func main() {
	http.HandleFunc("/", getCepHandler)
	http.ListenAndServe(":8080", nil)
}

func getCepHandler(w http.ResponseWriter, r *http.Request) {
	// Check route
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	cepParam := r.URL.Query().Get("cep")
	if cepParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// validate cep param
	regexp, _ := regexp.Compile(`[0-9]{5}-?[0-9]{3}`)
	if !regexp.MatchString(cepParam) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(cepParam) == 8 {
		cepParam = cepParam[:5] + "-" + cepParam[5:]
	}

	// create channel
	cepChannel1 := make(chan *CepData)
	cepChannel2 := make(chan *CepData)

	// get cep by api
	go getCepByApiCep(cepParam, cepChannel1)
	go getCepByViaCep(cepParam, cepChannel2)

	// wait for response
	for {
		select {
		case cep := <-cepChannel1:
			writeOnConsole(cep)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(cep)
			return
		case cep := <-cepChannel2:
			writeOnConsole(cep)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(cep)
			return
		case <-time.After(1 * time.Second):
			fmt.Println("Timeout")
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}
	}
}

func writeOnConsole(cep *CepData) {
	fmt.Println("=====================================")
	fmt.Println("CEP:", cep.Cep)
	fmt.Println("Logradouro:", cep.Logradouro)
	fmt.Println("Bairro:", cep.Bairro)
	fmt.Println("Localidade:", cep.Localidade)
	fmt.Println("UF:", cep.Uf)
	fmt.Println("API:", cep.Api)
	fmt.Println("=====================================")
}

func getCepByViaCep(cepParam string, cepChannel chan *CepData) {
	resp, error := http.Get("https://viacep.com.br/ws/" + cepParam + "/json/")

	// For test purpose
	// resp, error := http.Get("http://localhost:3001/viacep")
	// time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)

	if error != nil {
		error = fmt.Errorf("Error on get cep by viacep: %v", error)
	}
	defer resp.Body.Close()
	body, error := ioutil.ReadAll(resp.Body)
	if error != nil {
		error = fmt.Errorf("Error on read response body: %v", error)
	}
	var cep ViaCEP
	error = json.Unmarshal(body, &cep)
	if error != nil {
		error = fmt.Errorf("Error on unmarshal cep: %v", error)
	}

	cepChannel <- &CepData{
		Cep:        cep.Cep,
		Logradouro: cep.Logradouro,
		Bairro:     cep.Bairro,
		Localidade: cep.Localidade,
		Uf:         cep.Uf,
		Api:        "ViaCEP",
	}
}

func getCepByApiCep(cepParam string, cepChannel chan *CepData) {
	resp, error := http.Get("https://cdn.apicep.com/file/apicep/" + cepParam + ".json")

	// For test purpose
	// resp, error := http.Get("http://localhost:3001/apicep")
	// time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)

	if error != nil {
		error = fmt.Errorf("Error on get cep by apicep: %v", error)
	}
	defer resp.Body.Close()
	body, error := ioutil.ReadAll(resp.Body)
	if error != nil {
		error = fmt.Errorf("Error on read response body: %v", error)
	}
	var cep ApiCEP
	error = json.Unmarshal(body, &cep)
	if error != nil {
		error = fmt.Errorf("Error on unmarshal cep: %v", error)
	}

	cepChannel <- &CepData{
		Cep:        cep.Cep,
		Logradouro: cep.Logradouro,
		Bairro:     cep.Bairro,
		Localidade: cep.Localidade,
		Uf:         cep.Uf,
		Api:        "ApiCep",
	}
}
