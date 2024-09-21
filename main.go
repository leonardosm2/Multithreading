package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
)

func buscarDadosEndereco[T BrasilApiRetorno | ViaCepRetorno](ctx context.Context, url string, ch chan<- T) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var dados T
	err = json.Unmarshal(body, &dados)
	if err != nil {
		panic(err)
	}

	ch <- dados
}

func main() {
	args := os.Args

	cep := "01153000"
	if len(args) > 1 {
		cep = args[1]
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chBrasilApi := make(chan BrasilApiRetorno)
	chViaCep := make(chan ViaCepRetorno)

	go func() {
		url := "https://brasilapi.com.br/api/cep/v1/" + cep
		//time.Sleep(time.Second * 2)
		buscarDadosEndereco(ctx, url, chBrasilApi)
	}()

	go func() {
		url := "http://viacep.com.br/ws/" + cep + "/json/"
		//time.Sleep(time.Second * 2)
		buscarDadosEndereco(ctx, url, chViaCep)
	}()

	select {
	case dados := <-chBrasilApi:
		cancel()
		fmt.Printf(
			"Dados recebidos de BrasilApi.\nCEP: %s\nLogradouro: %s\nBairro: %s\nCidade: %s\nUF: %s\n",
			dados.Cep, dados.Logradouro, dados.Bairro, dados.Cidade, dados.UF,
		)

	case dados := <-chViaCep:
		cancel()
		fmt.Printf(
			"Dados recebidos de ViaCep.\nCEP: %s\nLogradouro: %s\nBairro: %s\nCidade: %s\nUF: %s\n",
			dados.Cep, dados.Logradouro, dados.Bairro, dados.Cidade, dados.UF,
		)

	case <-time.After(time.Second):
		cancel()
		println("timeout")
	}
}
