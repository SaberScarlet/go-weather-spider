package main

import (
	"os"
	"bufio"
	"io"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"mime/multipart"
	"bytes"
	"fmt"

	"strings"
)
type P map[string]interface{}
var Cityname = P{"beijing": "北京"}
var Cityweater string
var tablename=P{"城市":"citynm","省份":"provice","天气":"weather","气温℃":"temp","湿度(%)":"humidity","风向":"wind","风速(级别)":"省份","空气质量级别":"winp","PM2.5":"api","时间":"update"}
const  (
	weathername  = "城市,省份,天气,气温℃,湿度(%),风向,风速(级别),空气质量级别,PM2.5,时间\n"
	upurl_path string = "https://www.datahunter.cn/api/pub"
	upload_url string  ="https://www.datahunter.cn/api/upload"
)
func Get(url string) (content string, statusCode int) {
	resp, err1 := http.Get(url)
	if err1 != nil {
		statusCode = -100
		return
	}
	defer resp.Body.Close()
	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		statusCode = -200
		return
	}
	statusCode = resp.StatusCode
	content = string(data)
	return
}
func GetthType(name string)(nametype string){
	if strings.Contains(name, "风速(级别)")||strings.Contains(name, "湿度")||strings.Contains(name, "PM2.5")||strings.Contains(name, "气温(℃)"){
		return  "int"
	}else  {
		return "text"
	}

}
func JsonDecode(b []byte) (p *P) {
	p = &P{}
	err := json.Unmarshal(b, p)
	if err != nil {
		Error("JsonDecode", string(b), err)
	}
	return
}
func Error(i string, string string, i2 error) {

}
func Tocsv(csvpath string,text string )(err string){

	outputFile, outputError := os.OpenFile(csvpath, os.O_WRONLY|os.O_CREATE, 0666)
	if outputError !=nil {
		err = "nil"
		fmt.Println("发生错误打开文件的时候")
		return
	}

	defer outputFile.Close()
	outputWriter := bufio.NewWriter(outputFile)
	_,oa:=outputWriter.WriteString(text)
	if oa != nil{
		fmt.Println("发生错误打开文件的时候")
		err = "nil"
		return
	}
	outputWriter.Flush()
	err="ok"
	return

}
func Uplaodcsv(name string,Filepath string,key string,mode string,fmt1 string,th string)(err interface{}){
	fmt.Println(Filepath)
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	//关键的一步操作
	fileWriter, err := bodyWriter.CreateFormFile("bin", Filepath)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}
	//打开文件句柄操作
	fh, err := os.Open(Filepath)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	defer fh.Close()
	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		fmt.Println("第一步失败")
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(upload_url, contentType, bodyBuf)
	if err != nil {
		fmt.Println("第二步失败")
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("第三步失败")
		return err
	}
	fmt.Println(string(resp_body))
	fmt.Println(resp.Status)



	p:=*JsonDecode([]byte(string(resp_body)))

	url:=p["msg"].(map[string]interface{})["url"].(string)
	Upurl(name,url,key,mode,fmt1,th)

	return nil
}
func Upurl(name string,url string,key string,mode string, fmt1 string,th string)(err interface{}){
	client := &http.Client{}
	var params =map[string]string{"name":name,"url":url,"key":key,"mode":mode,"fmt":fmt1,"th":th}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, val := range params {
		fmt.Println(key)
		fmt.Println(val)
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return  err
	}
	request, err := http.NewRequest("POST", upurl_path, body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(request)
	fmt.Println(resp)
	return  err
}
func main() {
	for k,_:=range Cityname{
		s, statusCode := Get("http://api.k780.com:88/?app=weather.history&weaid=" + k + "&date=" + "2017-09-09" + "&appkey=23789&sign=abe1ba69c5f65c3fd1d95c535a5f7ed4&format=json")
		if statusCode != 200 {
			return
		}
		p:=*JsonDecode([]byte(s))
		a:=p["result"].([]interface {})
		for _,v:=range a{
			fmt.Println(v)
			b:=v.(map[string]interface{})
			Cityweater=Cityweater+b["citynm"].(string)+","
			Cityweater=Cityweater+Cityname[b["cityno"].(string)].(string)+","
			Cityweater=Cityweater+b["weather"].(string)+","
			Cityweater=Cityweater+b["citynm"].(string)+","
			Cityweater=Cityweater+b["temp"].(string)+","
			Cityweater=Cityweater+b["humidity"].(string)+","
			Cityweater=Cityweater+b["wind"].(string)+","
			Cityweater=Cityweater+b["winp"].(string)+","
			Cityweater=Cityweater+b["aqi"].(string)+","
			Cityweater=Cityweater+b["uptime"].(string)+"\n"
		}
	}
	tocsverr:=Tocsv("F:\\gopachong\\天气数据.csv",weathername+Cityweater)
	if tocsverr =="nil"{
		fmt.Println("csv输出失败")
	}else {
		fmt.Println("csv成功生成,接下来上传csv")
		slice:=strings.Split( strings.Trim(weathername, "\n"),",")
		namep:= []P{}

		for _,v:=range slice{
			nametype:=GetthType(v)
			p:=P{}
			p["o"]=tablename[v]
			p["n"]=v
			p["type"]=nametype
			namep=append(namep,p)
		}

		byte,_:=json.Marshal(namep)


		Uplaodcsv("测试","F:\\gopachong\\天气数据.csv","mrocker","2","gdp",string(byte))




	}

}




