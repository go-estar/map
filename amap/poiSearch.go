package amap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-estar/logger"
	"github.com/go-resty/resty/v2"
)

type PoiInfo struct {
	Parent   string `json:"parent"`
	Address  string `json:"address"`
	Distance string `json:"distance"`
	Pcode    string `json:"pcode"`
	Adcode   string `json:"adcode"`
	Pname    string `json:"pname"`
	Cityname string `json:"cityname"`
	Type     string `json:"type"`
	Typecode string `json:"typecode"`
	Adname   string `json:"adname"`
	Citycode string `json:"citycode"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Id       string `json:"id"`
}

type PoiSearchReq struct {
	Types  string `json:"types"`
	Radius string `json:"radius"`
	Lng    string `json:"lng"`
	Lat    string `json:"lat"`
}

type PoiSearchRes struct {
	Count    string     `json:"count"`
	Infocode string     `json:"infocode"`
	Pois     []*PoiInfo `json:"pois"`
	Status   string     `json:"status"`
	Info     string     `json:"info"`
}

func (a *AMap) PoiSearch(storeNo string,req *PoiSearchReq) (poi []*PoiInfo, err error) {
	var response string
	defer func() {
		a.Logger.Info(storeNo,
			logger.NewField("lng", req.Lng),
			logger.NewField("lat", req.Lat),
			logger.NewField("response",response),
			logger.NewField("error", err))
	}()

	if err := a.Limiter.Wait(context.Background()); err != nil {
		return nil, err
	}

	queryParams := map[string]string{
		"types":    req.Types,
		"radius":   req.Radius,
		"key":      a.Key,
		"location": req.Lng + "," + req.Lat,
	}
	resp, err := resty.New().R().SetQueryParams(queryParams).Execute("GET", "https://restapi.amap.com/v5/place/around")
	if err != nil {
		return nil, err
	}
	response = resp.String()
	var res = PoiSearchRes{}
	if err := json.Unmarshal(resp.Body(), &res); err != nil {
		return nil, err
	}
	if res.Status != "1" {
		return nil, errors.New(fmt.Sprintf("%s(%s)", res.Info, res.Infocode))
	}

	if len(res.Pois) == 0 {
		return nil, errors.New("未查询到Poi")
	}

	return res.Pois, nil
}
