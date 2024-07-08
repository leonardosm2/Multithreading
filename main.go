package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type (
	BrasilApiRetorno struct {
		Cep        string `json:"cep"`
		Logradouro string `json:"street"`
		Bairro     string `json:"neighborhood"`
		Cidade     string `json:"city"`
		UF         string `json:"state"`
	}

	ViaCepRetorno struct {
		Cep        string `json:"cep"`
		Logradouro string `json:"logradouro"`
		Bairro     string `json:"bairro"`
		Cidade     string `json:"localidade"`
		UF         string `json:"uf"`
	}

	DadosEndereco struct {
		Cep        string
		Logradouro string
		Bairro     string
		Cidade     string
		UF         string
	}

	Resultado struct {
		dados  DadosEndereco
		origem string
		err    error
	}
)

func buscarBrasilAPI(ctx context.Context, cep string, ch chan<- Resultado) {
	url := "https://brasilapi.com.br/api/cep/v1/" + cep

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- Resultado{dados: DadosEndereco{}, origem: "brasilapi", err: err}
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- Resultado{dados: DadosEndereco{}, origem: "brasilapi", err: err}
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- Resultado{dados: DadosEndereco{}, origem: "brasilapi", err: err}
		return
	}

	var brasilApiRetorno BrasilApiRetorno
	err = json.Unmarshal(body, &brasilApiRetorno)
	if err != nil {
		ch <- Resultado{dados: DadosEndereco{}, origem: "brasilapi", err: err}
		return
	}

	dadosEndereco := DadosEndereco{
		Cep:        brasilApiRetorno.Cep,
		Logradouro: brasilApiRetorno.Logradouro,
		Bairro:     brasilApiRetorno.Bairro,
		Cidade:     brasilApiRetorno.Cidade,
		UF:         brasilApiRetorno.UF,
	}

	ch <- Resultado{dados: dadosEndereco, origem: "brasilapi", err: nil}
}

func buscarViaCEP(ctx context.Context, cep string, ch chan<- Resultado) {
	url := "http://viacep.com.br/ws/" + cep + "/json/"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- Resultado{dados: DadosEndereco{}, origem: "viacep", err: err}
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- Resultado{dados: DadosEndereco{}, origem: "viacep", err: err}
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- Resultado{dados: DadosEndereco{}, origem: "viacep", err: err}
		return
	}

	var viaCepRetorno ViaCepRetorno
	err = json.Unmarshal(body, &viaCepRetorno)
	if err != nil {
		ch <- Resultado{dados: DadosEndereco{}, origem: "viacep", err: err}
		return
	}

	dadosEndereco := DadosEndereco{
		Cep:        viaCepRetorno.Cep,
		Logradouro: viaCepRetorno.Logradouro,
		Bairro:     viaCepRetorno.Bairro,
		Cidade:     viaCepRetorno.Cidade,
		UF:         viaCepRetorno.UF,
	}

	ch <- Resultado{dados: dadosEndereco, origem: "viacep", err: nil}
}

func main() {
	cep := "07085310"

	ch := make(chan Resultado, 2)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go buscarBrasilAPI(ctx, cep, ch)
	go buscarViaCEP(ctx, cep, ch)

	select {
	case res := <-ch:
		if res.err != nil {
			fmt.Println("Error:", res.err)
		} else {
			fmt.Printf("Dados do endereço: %+v\nOrigem: %s\n", res.dados, res.origem)
		}
	case <-ctx.Done():
		fmt.Println("Error: Timeout")
	}
}
