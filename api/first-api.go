package api

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
)

type FirstApi struct {
	url         string
	method      string
	contentType ContentType
}

func (rcv FirstApi) GetURL() string {
	return rcv.url
}

type firstApiRequest struct {
	SourceAddress string    `json:"contact address"`
	DestAddress   string    `json:"warehouse address"`
	BoxDimensions []float64 `json:"package dimensions"`
}

type firstApiResponse struct {
	Total interface{} `json:"total" xml:"total"`
}

func NewFirstApi() Resource {
	return &FirstApi{
		url:         "http://localhost:1111/",
		method:      "POST",
		contentType: ApplicationJson,
	}
}

func (rcv *FirstApi) GetAmount(data *Input, client http.Client) (float64, error) {
	body, err := json.Marshal(rcv.encodeFirstApiRequest(data))
	if err != nil {
		return 0, fmt.Errorf("marshall body err: %s", err)
	}

	req, err := http.NewRequest(rcv.method, rcv.url, bytes.NewBuffer(body))
	if err != nil {
		return 0, fmt.Errorf("new request err: %s", err)
	}

	req.Header.Add("Accept", rcv.contentType.String())

	rsp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do request err: %s", err)
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("response status code not OK: status: %s", rsp.Status)
	}

	var responseData firstApiResponse
	switch rcv.contentType {
	case ApplicationJson:
		if err := json.NewDecoder(rsp.Body).Decode(&responseData); err != nil {
			return 0, fmt.Errorf("json decode response err: %s", err)
		}
	case ApplicationXML:
		if err := xml.NewDecoder(rsp.Body).Decode(&responseData); err != nil {
			return 0, fmt.Errorf("xml decode response err: %s", err)
		}
	}

	var amount float64
	switch t := responseData.Total.(type) {
	case float64:
		amount = responseData.Total.(float64)
	case string:
		amount, err = strconv.ParseFloat(responseData.Total.(string), 64)
		if err != nil {
			return 0, fmt.Errorf("can't parse float | rsp:%s | err: %s", responseData.Total, err)
		} else if amount == 0 {
			return 0, fmt.Errorf("got amount: 0, so will not count as a real offer")
		}
	default:
		return 0, fmt.Errorf("not supported value type: %T", t)
	}

	return amount, nil
}

func (rcv FirstApi) encodeFirstApiRequest(data *Input) firstApiRequest {
	return firstApiRequest{
		SourceAddress: data.SourceAddress,
		DestAddress:   data.DestAddress,
		BoxDimensions: data.BoxDimensions,
	}
}
