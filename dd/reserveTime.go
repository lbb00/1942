package dd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"
)

type ReserveTime struct {
	StartTimestamp int    `json:"start_timestamp"`
	EndTimestamp   int    `json:"end_timestamp"`
	SelectMsg      string `json:"select_msg"`
}

func (s *DingdongSession) GetMultiReserveTime() (error, []ReserveTime) {
	urlPath := "https://maicai.api.ddxq.mobi/order/getMultiReserveTime"
	var products []map[string]interface{}
	for _, product := range s.Order.Products {
		prod := map[string]interface{}{
			"id": product.Id,
			// 这些字段暂时不需要
			// "total_money":          product.TotalPrice,
			// "total_origin_money":   product.OriginPrice,
			// "count":                product.Count,
			// "price":                product.Price,
			// "instant_rebate_money": "0.00",
			// "origin_price":         product.OriginPrice,
		}
		products = append(products, prod)
	}

	productsJson, _ := json.Marshal([][]map[string]interface{}{
		products,
	})

	data := url.Values{}
	data.Add("station_id", s.Address.StationId)
	data.Add("city_number", s.Address.CityNumber)
	data.Add("api_version", "9.49.0")
	data.Add("app_version", "2.81.0")
	data.Add("applet_source", "")
	data.Add("app_client_id", "4")
	data.Add("h5_source", "")
	data.Add("sharer_uid", "")
	data.Add("s_id", "")
	data.Add("openid", "")
	data.Add("group_config_id", "")
	data.Add("products", string(productsJson))
	data.Add("isBridge", "false")
	data.Add("address_id", s.Address.Id)

	req, _ := http.NewRequest("POST", urlPath, strings.NewReader(data.Encode()))
	req.Header.Set("ddmc-city-number", s.Address.CityNumber)
	req.Header.Set("ddmc-station-id", s.Address.StationId)
	BindCommonHeader(req)

	resp, err := s.Client.Do(req)
	if err != nil {
		return err, nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		var reserveTimeList []ReserveTime
		result := gjson.Parse(string(body))
		for _, reserveTimeInfo := range result.Get("data.0.time.0.times").Array() {
			if reserveTimeInfo.Get("disableType").Num == 0 {
				reserveTime := ReserveTime{
					StartTimestamp: int(reserveTimeInfo.Get("start_timestamp").Num),
					EndTimestamp:   int(reserveTimeInfo.Get("end_timestamp").Num),
					SelectMsg:      reserveTimeInfo.Get("select_msg").Str,
				}
				reserveTimeList = append(reserveTimeList, reserveTime)
			}
		}
		return nil, reserveTimeList
	} else {
		return fmt.Errorf("[%v] %s", resp.StatusCode, body), nil
	}
}
