package main

import (
	"errors"
	"log"
	"sync"

	ucapi "github.com/enbility/eebus-go/usecases/api"
	shipapi "github.com/enbility/ship-go/api"
	"github.com/gorilla/websocket"
)

type WebsocketClient struct {
	websocket *websocket.Conn
	mutex     sync.Mutex
	mutex2    sync.Mutex
}

func (websocketClient *WebsocketClient) sendMessage(msg interface{}) error {
	if websocketClient.websocket == nil {
		return errors.New("no frontend connected")
	}

	websocketClient.mutex.Lock()
	defer websocketClient.mutex.Unlock()

	err := websocketClient.websocket.WriteJSON(msg)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (websocketClient *WebsocketClient) sendNotification(ski string, messageType int, uc string) error {
	answer := Message{
		SKI:     ski,
		Type:    messageType,
		UseCase: uc}

	return websocketClient.sendMessage(answer)
}

func (websocketClient *WebsocketClient) sendText(messageType int, text string) error {
	answer := Message{
		Type: messageType,
		Text: text}

	return websocketClient.sendMessage(answer)
}

func (websocketClient *WebsocketClient) sendValue(ski string, messageType int, useCase string, value float64) error {
	answer := Message{
		SKI:     ski,
		Type:    messageType,
		Value:   value,
		UseCase: useCase}

	return websocketClient.sendMessage(answer)
}

func (websocketClient *WebsocketClient) sendValueArr(ski string, messageType int, useCase string, values []float64) error {
	answer := Message{
		SKI:     ski,
		Type:    messageType,
		Values:  values,
		UseCase: useCase}

	return websocketClient.sendMessage(answer)
}

func (websocketClient *WebsocketClient) sendLimit(ski string, messageType int, useCase string, limit ucapi.LoadLimit) error {
	answer := Message{
		SKI:     ski,
		Type:    messageType,
		Limit:   limit,
		UseCase: useCase}

	return websocketClient.sendMessage(answer)
}

func (websocketClient *WebsocketClient) sendServiceList(messageType int, services []shipapi.RemoteService) error {
	answer := Message{
		Type:        messageType,
		ServiceList: services}

	return websocketClient.sendMessage(answer)
}

func (websocketClient *WebsocketClient) sendEntityInfo(messageType int, remoteInfos map[string]RemoteInfo) error {
	websocketClient.mutex2.Lock()
	defer websocketClient.mutex2.Unlock()

	entityInfos := []EntityInfo{}

	for _, remoteInfo := range remoteInfos {
		device := remoteInfo.Device
		if device != nil {
			for _, entity := range device.Entities() {
				features := []string{}

				for _, f := range entity.Features() {
					features = append(features, f.String()+", "+string(f.Role()))
				}

				info := EntityInfo{
					Address:  entity.Address().String(),
					Name:     string(entity.EntityType()),
					SKI:      device.Ski(),
					Type:     string(*device.DeviceType()),
					Features: features}

				entityInfos = append(entityInfos, info)
			}
		}
	}

	answer := Message{
		Type:        messageType,
		EntityInfos: entityInfos}

	return websocketClient.sendMessage(answer)
}

func (websocketClient *WebsocketClient) sendUseCaseInfo(messageType int, useCaseInfos map[string][]UseCaseInfo) error {
	websocketClient.mutex2.Lock()
	defer websocketClient.mutex2.Unlock()

	answer := Message{
		Type:         messageType,
		UseCaseInfos: useCaseInfos}

	return websocketClient.sendMessage(answer)
}
