package gohessian

//
//import (
//	"bytes"
//	"log"
//	"testing"
//)
//
//const (
//	MATH_H_URL = "http://localhost:8080/HessianTest/math" //整数四则运算hessian测试接口
//	DT_H_URL   = "http://localhost:8080/HessianTest/dt"   //数据类型测试结果,传入参数，返回该参数的(响应)编码结果
//)
//
//func Test_request_http_post(t *testing.T) {
//	t.SkipNow()
//	log.Println("Test_request")
//	data := bytes.NewBuffer([]byte{0, 1, 3, 4})
//	rb, _ := httpPost(DT_H_URL, bytes.NewReader(data.Bytes()))
//	log.Println(rb)
//	log.Println(string(rb))
//}
//
////整数 数学运算测试
//func Test_request_int_math(t *testing.T) {
//	mathClient := NewClient(MATH_H_URL, "")
//	//Request(H_URL, "add", 100, 200, 101.5, true, false, []byte{1, 2, 3, 5})
//	mathClient.Invoke("add", 100, 200)
//	mathClient.Invoke("sub", 100, 200)
//	mathClient.Invoke("mult", 100, 200)
//	mathClient.Invoke("div", 200, 50)
//}
//
////数据类型测试
//func Test_request_data_type(t *testing.T) {
//	dtClient := NewClient(DT_H_URL, "")
//	dtClient.Invoke("dataBytes", []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
//	dtClient.Invoke("dataBoolean", true)
//	dtClient.Invoke("dataBoolean", false)
//	dtClient.Invoke("dataDouble", 1989.0604)
//
//	list := []Any{100, 10.001, "不厌其烦", []byte{0, 2, 4, 6, 8, 10}, true, nil, false}
//	dtClient.Invoke("dataList", list)
//
//	var hmap = make(map[Any]Any)
//	hmap["你好"] = "哈哈哈"
//	hmap[100] = "嘿嘿"
//	hmap[100.1010] = 101910
//	hmap[true] = true
//	hmap[false] = true
//	dtClient.Invoke("dataMap", hmap)
//
//	dtClient.Invoke("dataMapNoParam")
//
//	dtClient.Invoke("dataNull")
//
//	dtClient.Invoke("dataString", "_BEGIN_兔兔你小姨子_END_")
//
//	dtClient.Invoke("dataInt", 1000)
//
//}
//
////异常测试,服务器已经抛出异常，但客户端看到的是200和空响应
////curl -vvv --data-binary "c\x00\x01m\x00\adataIntz" -H "Content-Type: application/binary" http://localhost:8080/HessianTest/dt
////curl -vvv --data-binary "c\x00\x01m\x00\x0EthorwExceptionz" -H "Content-Type: application/binary" http://localhost:8080/HessianTest/dt
//func Test_request_exception(t *testing.T) {
//	dtClient := NewClient(DT_H_URL, "")
//	// Request(DT_H_URL,"dataInt")
//	dtClient.Invoke("thorwException")
//}
