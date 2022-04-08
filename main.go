package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"food/dd"
	"food/logger"
	"food/tasks"
)

type Config struct {
	Cookie string
	BarkId string
	Uid    string
}

var (
	cartSleepTime    = time.Millisecond * 5000
	orderSleepTime   = time.Millisecond * 5000
	reserveSleepTime = time.Millisecond * 5000
)

var refreshSleepTime = time.Millisecond * 5000

var app *tasks.Task
var session dd.DingdongSession

var reserveTimeList []dd.ReserveTime

func getReserveTime(task *tasks.Task) {

	for {
		select {
		case <-task.Ctx.Done():
			return
		default:
			logger.Log.Infof("获取可预约时间\n", time.Now().Format("15:04:05"))
			err, list := session.GetMultiReserveTime()
			if err != nil {
				logger.Log.Warnf("获取预约时间失败")
				time.Sleep(reserveSleepTime)
			} else {
				reserveTimeList = list
				count := len(reserveTimeList)
				logger.Log.Infof("共 %v 个时间段可预约", count)
				// 请求成功后续就不用请求那么频繁了
				time.Sleep(refreshSleepTime)
			}
		}
	}
}

func createOrder(task *tasks.Task) {
	// 这里直接使用全局的可预约时间，大概率不需要考虑商品发生变化后可预约时间也发生变化了

	checkOrderSuccess := false
	for {
		select {
		case <-task.Ctx.Done():
			return
		default:
			if !checkOrderSuccess {
				logger.Log.Infof("获取订单金额中...")
				err := session.CheckOrder()
				if err != nil {
					fmt.Println(err)
					time.Sleep(orderSleepTime)
					continue
				}
				logger.Log.Infof("订单总金额：%v\n", session.Order.Price)
				checkOrderSuccess = true
				session.GeneratePackageOrder()
			}

			// 并发！！同时向所有未约满的时间提交
			// 疫情期间大概每天4个时间段可以选，非疫情就多了，不能这么玩咯
			if len(reserveTimeList) > 0 {
				createTask := tasks.NewTask(context.TODO())
				createSuccess := false
				go func() {
					wg := sync.WaitGroup{}
					for _, reserveTime := range reserveTimeList {
						wg.Add(1)
						go func(reserveTime dd.ReserveTime, s dd.DingdongSession) {
							select {
							case <-createTask.Ctx.Done():
								wg.Done()
							default:
								logger.Log.Infof("提交订单中 预约时间 %s\n", reserveTime.SelectMsg)
								s.UpdatePackageOrder(reserveTime)
								err := s.AddNewOrder()
								switch err {
								case nil:
									createTask.Cancel()
									createSuccess = true
								case dd.TimeExpireErr:
									logger.Log.Warnf("%s 预约时间%s \n", err, reserveTime.SelectMsg)
								case dd.ProdInfoErr:
									logger.Log.Warn(err)
								default:
									logger.Log.Warn(err)
								}
								wg.Done()
							}

						}(reserveTime, session)
					}
					wg.Wait()
					createTask.Cancel()
				}()
				<-createTask.Ctx.Done()
				if createSuccess {
					logger.Log.Infof("抢购成功，请前往app付款！")
					task.Cancel()
					app.Cancel()
				}

				time.Sleep(orderSleepTime)
			}

		}
	}
}

func start() {
	var task *tasks.Task
	var cartHash string
	for {
		select {
		case <-app.Ctx.Done():
			fmt.Println("抢购完毕，程序退出")
			return
		default:

			if err := session.CheckCart(); err != nil {
				fmt.Println(err)
				time.Sleep(cartSleepTime)
				continue
			}

			if len(session.Cart.ProdList) == 0 {
				fmt.Println("购物车中无有效商品，请先前往app添加或勾选！")
				app.Cancel()
				return
			}
			session.Order.Products = session.Cart.ProdList
			hash, _ := json.Marshal(session.Cart)
			// 直接字符串比较吧，懒得hash了
			if string(hash) != cartHash {
				if task != nil {
					task.Cancel()
				}
				if cartHash != "" {
					fmt.Println("--- 购物车发生变化,执行新的任务 ---")
				}
				for index, prod := range session.Cart.ProdList {
					fmt.Printf("[%v] %s 数量：%v 总价：%s\n", index, prod.ProductName, prod.Count, prod.TotalPrice)
				}

				cartHash = string(hash)
				task = tasks.NewTask(context.TODO())

				go getReserveTime(task)
				go createOrder(task)
			}
			time.Sleep(refreshSleepTime)
		}

	}
}

func main() {
	logger.Init()

	app = tasks.NewTask(context.Background())

	session = dd.DingdongSession{}

    // 配置这里
	header := dd.CommonHeader{
		Cookie:    "",
		Uid:       "",
		DeviceId:  "",

        // 这两个可以不配置
		Longitude: "",
		Latitude:  "",
	}

	barkId := ""

	if err := session.InitSession(header, barkId); err != nil {
		fmt.Println(err)
		return
	}
	start()
}
