package address

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
)

type ViaCep struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Unidade     string `json:"unidade"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Estado      string `json:"estado"`
	Regiao      string `json:"regiao"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

func GetCep(cep string, ctx context.Context) (*ViaCep, error) {
	// Intrumenta o span para a chamada interna
	// Pega o tracer novamente (ou poderia ser passado como argumento)
	tracer := otel.Tracer("service-b")

	// Inicia um span filho, pois estamos usando o contexto do `helloHandlerSpan`
	_, span := tracer.Start(ctx, "GetLocationByCepSpan")
	defer span.End()

	time.Sleep(2 * time.Second) // Simula algum processamento

	// Desabilitar a verificação do certificado SSL
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	_, err := checkCep(cep)
	if err != nil {
		fmt.Errorf("Error: %w", err)
	}

	resp, error := http.Get("https://viacep.com.br/ws/" + cep + "/json/")
	if error != nil {
		return nil, error
	}
	defer resp.Body.Close()

	body, error := io.ReadAll(resp.Body)
	if error != nil {
		return nil, error
	}
	// fmt.Println("Address resp: ", string(body))

	var c ViaCep
	error = json.Unmarshal(body, &c)
	if error != nil {
		return nil, error
	}

	return &c, nil
}

func checkCep(cep string) (bool, error) {
	if len(cep) != 8 {
		return false, fmt.Errorf("invalid zipcode")
	}
	return true, nil
}
