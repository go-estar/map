package amap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-estar/logger"
	"github.com/go-estar/types/fieldUtil"
	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"
)

type GeocodeRes struct {
	Status    string `json:"status"`
	Regeocode struct {
		AddressComponent struct {
			Province interface{} `json:"province"`
			City     interface{} `json:"city"`
			District interface{} `json:"district"`
			Township interface{} `json:"township"`
			Adcode   interface{} `json:"adcode"`
		} `json:"addressComponent"`
		FormattedAddress interface{} `json:"formatted_address"`
	} `json:"regeocode"`
	Info     string `json:"info"`
	Infocode string `json:"infocode"`
}

type AddressInfo struct {
	Province     string `json:"province"`     // 省份
	ProvinceCode string `json:"provinceCode"` // 省份编码
	City         string `json:"city"`         // 城市
	CityCode     string `json:"cityCode"`     // 城市编码
	District     string `json:"district"`     // 区
	DistrictCode string `json:"districtCode"` // 区编码
	Address      string `json:"address"`      // 地址
}

func New(key string, logger2 logger.Logger) *AMap {
	if key == "" {
		panic("amap key must set")
	}
	if logger2 == nil {
		panic("amap logger must set")
	}
	return &AMap{
		Key:     key,
		Limiter: rate.NewLimiter(1, 1),
		Logger:  logger2,
	}
}

type AMap struct {
	Key string
	*rate.Limiter
	logger.Logger
}

func (a *AMap) Geocode(storeNo, lng, lat string) (addressInfo *AddressInfo, err error) {
	defer func() {
		a.Logger.Info(storeNo,
			logger.NewField("lng", lng),
			logger.NewField("lat", lat),
			logger.NewField("error", err))
	}()

	if err := a.Limiter.Wait(context.Background()); err != nil {
		return nil, err
	}

	queryParams := map[string]string{
		"output":     "JSON",
		"extensions": "base",
		"key":        a.Key,
		"location":   lng + "," + lat,
	}
	resp, err := resty.New().R().SetQueryParams(queryParams).Execute("GET", "https://restapi.amap.com/v3/geocode/regeo")
	if err != nil {
		return nil, err
	}
	var res = GeocodeRes{}
	if err := json.Unmarshal(resp.Body(), &res); err != nil {
		return nil, err
	}
	if res.Status != "1" {
		return nil, errors.New(fmt.Sprintf("%s(%s)", res.Info, res.Infocode))
	}

	districtCode := fmt.Sprintf("%v", res.Regeocode.AddressComponent.Adcode)
	if len(districtCode) != 6 {
		return nil, errors.New("地址编码有误")
	}

	province := fmt.Sprintf("%v", res.Regeocode.AddressComponent.Province)
	city := ""
	if !fieldUtil.IsEmpty(res.Regeocode.AddressComponent.City) {
		city = fmt.Sprintf("%v", res.Regeocode.AddressComponent.City)
	} else {
		city = province
	}

	district := ""
	if !fieldUtil.IsEmpty(res.Regeocode.AddressComponent.District) {
		district = fmt.Sprintf("%v", res.Regeocode.AddressComponent.District)
	} else {
		district = fmt.Sprintf("%v", res.Regeocode.AddressComponent.Township)
	}
	if province == "" || city == "" || district == "" {
		return nil, errors.New("地址行政区有误")
	}

	return &AddressInfo{
		Province:     province,
		ProvinceCode: districtCode[0:2],
		City:         city,
		CityCode:     districtCode[0:4],
		District:     district,
		DistrictCode: districtCode,
		Address:      fmt.Sprintf("%v", res.Regeocode.FormattedAddress),
	}, nil
}
