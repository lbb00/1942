package dd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/tidwall/gjson"
)

type Product struct {
	Id               string                   `json:"id"`
	ProductName      string                   `json:"-"`
	Price            string                   `json:"price"`
	Count            int                      `json:"count"`
	Sizes            []map[string]interface{} `json:"sizes"`
	TotalPrice       string                   `json:"total_money"`
	OriginPrice      string                   `json:"origin_price"`
	TotalOriginPrice string                   `json:"total_origin_money"`
}

func parseProduct(productMap gjson.Result) Product {
	var sizes []map[string]interface{}
	for _, size := range productMap.Get("sizes").Array() {
		sizes = append(sizes, size.Value().(map[string]interface{}))
	}
	product := Product{
		Id:          productMap.Get("id").Str,
		ProductName: productMap.Get("product_name").Str,
		Price:       productMap.Get("price").Str,
		Count:       int(productMap.Get("count").Num),
		TotalPrice:  productMap.Get("total_price").Str,
		OriginPrice: productMap.Get("origin_price").Str,
		Sizes:       sizes,
	}
	return product
}

type Cart struct {
	ProdList        []Product `json:"effective_products"`
	ParentOrderSign string    `json:"parent_order_sign"`
}

func (s *DingdongSession) GetEffProd(result gjson.Result) error {
	var effProducts []Product
	effective := result.Get("data.product.effective").Array()
	for _, effProductMap := range effective {
		for _, productMap := range effProductMap.Get("products").Array() {
			product := parseProduct(productMap)
			effProducts = append(effProducts, product)
		}
	}
	s.Cart = Cart{
		ProdList:        effProducts,
		ParentOrderSign: result.Get("data.parent_order_info.parent_order_sign").Str,
	}
	return nil
}

func (s *DingdongSession) GetCheckProd(result gjson.Result) error {
	var products []Product
	orderProductList := result.Get("data.new_order_product_list").Array()
	for _, productList := range orderProductList {
		for _, productMap := range productList.Get("products").Array() {
			product := parseProduct(productMap)
			products = append(products, product)
		}
	}
	s.Cart = Cart{
		ProdList:        products,
		ParentOrderSign: result.Get("data.parent_order_info.parent_order_sign").Str,
	}
	return nil
}

func (s *DingdongSession) CheckCart() error {
	Url, _ := url.Parse("https://maicai.api.ddxq.mobi/cart/index")
	params := url.Values{}
	params.Set("station_id", s.Address.StationId)
	params.Set("city_number", s.Address.CityNumber)
	params.Set("api_version", "9.49.0")
	params.Set("app_version", "2.81.0")
	params.Set("applet_source", "")
	params.Set("app_client_id", "4")
	params.Set("h5_source", "")
	params.Set("sharer_uid", "")
	params.Set("s_id", "")
	params.Set("openid", "")
	params.Set("is_load", "1")
	params.Set("ab_config", "{\"key_onion\":\"D\",\"key_cart_discount_price\":\"C\"}")

	Url.RawQuery = params.Encode()
	urlPath := Url.String()
	req, _ := http.NewRequest("GET", urlPath, nil)

	req.Header.Set("ddmc-city-number", s.Address.CityNumber)
	req.Header.Set("ddmc-station-id", s.Address.StationId)
	BindCommonHeader(req)

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode == 200 {
		result := gjson.Parse(string(body))
		switch result.Get("code").Num {
		case 0:
			switch s.CartMode {
			case 1:
				return s.GetEffProd(result)
			case 2:
				return s.GetCheckProd(result)
			default:
				return errors.New("incorrect cart mode")
			}
		case -3000:
			return BusyErr
		default:
			return errors.New(string(body))
		}
	} else {
		return fmt.Errorf("[%v] %s", resp.StatusCode, body)
	}
}
