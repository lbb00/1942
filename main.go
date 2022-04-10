package main

import (
	"context"
	"encoding/json"
	"flag"
	"sync"
	"time"

	"food/config"
	"food/dd"
	"food/logger"
	"food/tasks"
)

// 捡漏用
var (
	cartSleepTime    = time.Millisecond * 3000
	orderSleepTime   = time.Millisecond * 3000
	reserveSleepTime = time.Millisecond * 3000
	refreshSleepTime = time.Millisecond * 5000
)

//  抢购用
// var (
// 	cartSleepTime    = time.Millisecond * 100
// 	orderSleepTime   = time.Millisecond * 100
// 	reserveSleepTime = time.Millisecond * 100
//  refreshSleepTime = time.Millisecond * 2000
// )

var (
	app             *tasks.Task
	session         dd.DingdongSession
	reserveTimeList []dd.ReserveTime
)

var (
	conf     config.Conf
	confName string
)

func init() {
	flag.StringVar(&confName, "conf", "", "")
}

func main() {
	logger.Init()

	flag.Parse()

	if confName == "" {
		logger.Log.Error("conf name is empty")
		return
	}
	if _, err := conf.GetConf(confName); err != nil {
		logger.Log.Error(err)
		return
	}

	app = tasks.NewTask(context.Background())

	session = dd.DingdongSession{}

	if err := session.InitSession(dd.CommonHeader{
		Cookie:   conf.Cookie,
		Uid:      conf.Uid,
		DeviceId: conf.DeviceId,

		Longitude: conf.Longitude,
		Latitude:  conf.Latitude,
	}, conf.BarkId); err != nil {
		logger.Log.Error(err)
		return
	}
	start()
}

func start() {
	var task *tasks.Task
	var cartHash string
	cartTime := time.Now()

	for {
		select {
		case <-app.Ctx.Done():
			logger.Log.Info("抢购完毕，程序退出")
			return
		default:
			if task != nil && time.Since(cartTime) < refreshSleepTime {
				time.Sleep(time.Millisecond * 50)
				continue
			}
			cartTime = time.Now()
			if err := session.CheckCart(); err != nil {
				logger.Log.Warnf("获取购物车失败 %s", err)
				time.Sleep(cartSleepTime)
				continue
			}

			if len(session.Cart.ProdList) == 0 {
				logger.Log.Error("购物车中无有效商品，请先前往app添加或勾选！")
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
					logger.Log.Info("--- 购物车发生变化,执行新的任务 ---")
				}
				for index, prod := range session.Cart.ProdList {
					logger.Log.Infof("[%v] %s 数量：%v 总价：%s", index, prod.ProductName, prod.Count, prod.TotalPrice)
				}

				cartHash = string(hash)
				task = tasks.NewTask(context.TODO())

				go getReserveTime(task)
				go createOrder(task)
			}

		}

	}
}
func getReserveTime(task *tasks.Task) {
	for {
		select {
		case <-task.Ctx.Done():
			return
		default:
			logger.Log.Info("获取可预约时间")
			err, list := session.GetMultiReserveTime()
			if err != nil {
				logger.Log.Warnf("获取预约时间失败 %s", err)
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
					logger.Log.Warnf("获取订单金额失败 %s", err)
					time.Sleep(orderSleepTime)
					continue
				}
				logger.Log.Infof("订单总金额：%v", session.Order.Price)
				checkOrderSuccess = true
				session.GeneratePackageOrder()
			}

			// 并发！！同时向所有未约满的时间提交
			// 疫情期间大概每天4个时间段可以选，非疫情就多了，不能这么玩咯
			if len(reserveTimeList) > 0 {
				createTask := tasks.NewTask(context.TODO())
				createSuccess := false
				hasProdErr := false
				go func() {
					wg := sync.WaitGroup{}
					for _, reserveTime := range reserveTimeList {
						wg.Add(1)
						go func(reserveTime dd.ReserveTime, s dd.DingdongSession) {
							select {
							case <-createTask.Ctx.Done():
								wg.Done()
							default:
								logger.Log.Infof("提交订单中 预约时间 %s", reserveTime.SelectMsg)
								s.UpdatePackageOrder(reserveTime)
								err := s.AddNewOrder()
								switch err {
								case nil:
									createTask.Cancel()
									createSuccess = true
								case dd.TimeExpireErr:
									hasProdErr = true
									logger.Log.Warnf("下单失败 预约时间%s %s", reserveTime.SelectMsg, err)
								case dd.ProdInfoErr:
									logger.Log.Warnf("下单失败 预约时间%s %s", reserveTime.SelectMsg, err)
								default:
									logger.Log.Warnf("下单失败 预约时间%s %s", reserveTime.SelectMsg, err)
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
					return
				}

				if hasProdErr {
					task.Cancel()
					return
				}

				time.Sleep(orderSleepTime)

			}

		}
	}
}
