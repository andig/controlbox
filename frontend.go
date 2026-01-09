package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"time"

	ucapi "github.com/enbility/eebus-go/usecases/api"
	shipapi "github.com/enbility/ship-go/api"
	spineapi "github.com/enbility/spine-go/api"
	"github.com/gorilla/websocket"
)

// web frontend

const (
	httpdPort int = 7080
)

const (
	Text                           = 0
	QRCode                         = 1
	Acknowledge                    = 2
	ServiceListChanged             = 3
	GetServiceList                 = 4
	SelectService                  = 5
	GetEntityInfos                 = 6
	GetUseCaseInfos                = 7
	GetAllData                     = 8
	SetConsumptionLimit            = 9
	GetConsumptionLimit            = 10
	SetProductionLimit             = 11
	GetProductionLimit             = 12
	SetConsumptionFailsafeValue    = 13
	GetConsumptionFailsafeValue    = 14
	SetConsumptionFailsafeDuration = 15
	GetConsumptionFailsafeDuration = 16
	SetProductionFailsafeValue     = 17
	GetProductionFailsafeValue     = 18
	SetProductionFailsafeDuration  = 19
	GetProductionFailsafeDuration  = 20
	GetConsumptionNominalMax       = 21
	GetProductionNominalMax        = 22
	GetConsumptionHeartbeat        = 23
	StopConsumptionHeartbeat       = 24
	StartConsumptionHeartbeat      = 25
	GetProductionHeartbeat         = 26
	StopProductionHeartbeat        = 27
	StartProductionHeartbeat       = 28
	GetPowerLimitationFactor       = 29
	GetPower                       = 30
	GetPowerPerPhase               = 31
	GetEnergyFeedIn                = 32
	GetEnergyConsumed              = 33
	GetCurrentPerPhase             = 34
	GetVoltagePerPhase             = 35
	GetFrequency                   = 36
)

type RemoteInfo struct {
	Service  shipapi.RemoteService
	Device   spineapi.DeviceRemoteInterface
	UseCases []string
}

type UseCaseInfo struct {
	Actor string
	Names []string
}

type EntityInfo struct {
	Address  string
	Name     string
	SKI      string
	Type     string
	Features []string
}

type Message struct {
	SKI          string
	Type         int
	Text         string
	Limit        ucapi.LoadLimit
	Value        float64
	Values       []float64
	ServiceList  []shipapi.RemoteService
	EntityInfos  []EntityInfo
	UseCaseInfos map[string][]UseCaseInfo
	UseCase      string
}

func readData(h *controlbox, entity spineapi.EntityRemoteInterface, ucs []string) {
	ski := entity.Device().Ski()

	if (ucs == nil || slices.Contains(ucs, "LPC")) && slices.Contains(h.remoteInfos[ski].UseCases, "LPC") {
		if currentLimit, err := h.uclpc.ConsumptionLimit(entity); err == nil {
			h.consumptionLimits = currentLimit

			frontend.sendLimit(ski, GetConsumptionLimit, "LPC", ucapi.LoadLimit{
				IsActive: currentLimit.IsActive,
				Duration: currentLimit.Duration / time.Second,
				Value:    currentLimit.Value})
		}

		if limit, err := h.uclpc.FailsafeConsumptionActivePowerLimit(entity); err == nil {
			h.consumptionFailsafeLimits.Value = limit

			frontend.sendValue(ski, GetConsumptionFailsafeValue, "LPC", limit)
		}

		if duration, err := h.uclpc.FailsafeDurationMinimum(entity); err == nil {
			h.consumptionFailsafeLimits.Duration = duration

			frontend.sendValue(ski, GetConsumptionFailsafeDuration, "LPC", float64(duration/time.Second))
		}

		if nominal, err := h.uclpc.ConsumptionNominalMax(entity); err == nil {
			h.consumptionNominalMax = nominal

			frontend.sendValue(ski, GetConsumptionNominalMax, "LPC", nominal)
		}
	}

	if (ucs == nil || slices.Contains(ucs, "LPP")) && slices.Contains(h.remoteInfos[ski].UseCases, "LPP") {
		if currentLimit, err := h.uclpp.ProductionLimit(entity); err == nil {
			h.productionLimits = currentLimit

			frontend.sendLimit(ski, GetProductionLimit, "LPP", ucapi.LoadLimit{
				IsActive: currentLimit.IsActive,
				Duration: currentLimit.Duration / time.Second,
				Value:    currentLimit.Value})
		}

		if limit, err := h.uclpp.FailsafeProductionActivePowerLimit(entity); err == nil {
			h.productionFailsafeLimits.Value = limit

			frontend.sendValue(ski, GetProductionFailsafeValue, "LPP", limit)
		}

		if duration, err := h.uclpp.FailsafeDurationMinimum(entity); err == nil {
			h.productionFailsafeLimits.Duration = duration

			frontend.sendValue(ski, GetProductionFailsafeDuration, "LPP", float64(duration/time.Second))
		}

		if nominal, err := h.uclpp.ProductionNominalMax(entity); err == nil {
			h.productionNominalMax = nominal

			frontend.sendValue(ski, GetProductionNominalMax, "LPP", nominal)
		}
	}
}

func sendData(h *controlbox, ski string, uc string) {
	switch uc {
	case "":
		// TODO
		// frontend.sendText(QRCode, h.myService.QRCodeText())
		frontend.sendText(QRCode, "")

	case "LPC":
		frontend.sendLimit(ski, GetConsumptionLimit, "LPC", ucapi.LoadLimit{
			IsActive: h.consumptionLimits.IsActive,
			Duration: h.consumptionLimits.Duration / time.Second,
			Value:    h.consumptionLimits.Value})

		frontend.sendValue(ski, GetConsumptionFailsafeValue, "LPC", h.consumptionFailsafeLimits.Value)

		frontend.sendValue(ski, GetConsumptionFailsafeDuration, "LPC", float64(h.consumptionFailsafeLimits.Duration/time.Second))

	case "LPP":
		frontend.sendLimit(ski, GetProductionLimit, "LPP", ucapi.LoadLimit{
			IsActive: h.productionLimits.IsActive,
			Duration: h.productionLimits.Duration / time.Second,
			Value:    h.productionLimits.Value})

		frontend.sendValue(ski, GetProductionFailsafeValue, "LPP", h.productionFailsafeLimits.Value)

		frontend.sendValue(ski, GetProductionFailsafeDuration, "LPP", float64(h.productionFailsafeLimits.Duration/time.Second))

	default:
		return
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// allow connection from any host
		return true
	},
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func serveWs(h *controlbox, w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	frontend = WebsocketClient{
		websocket: ws}

	log.Println("Client Connected")

	frontend.sendServiceList(GetServiceList, h.currentRemoteServices)

	sendData(h, "", "")

	reader(h, ws)
}

func reader(h *controlbox, ws *websocket.Conn) {
	for {
		// read in a message
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// print out that message for clarity
		//fmt.Println(string(p))

		data := Message{}
		json.Unmarshal([]byte(p), &data)

		switch data.Type {
		case GetServiceList:
			frontend.sendServiceList(GetServiceList, h.currentRemoteServices)
		case SelectService:
			remoteSki = data.Text

			info, exists := h.remoteInfos[remoteSki]
			if !exists {
				connected, exists2 := h.isConnected[remoteSki]
				if !exists2 || !connected {
					h.myService.RegisterRemoteSKI(remoteSki)
				}
			} else if info.Device != nil {
				for _, entity := range info.Device.Entities() {
					readData(h, entity, nil)
				}
			}
		case GetEntityInfos:
			if nil != h.remoteInfos {
				frontend.sendEntityInfo(GetEntityInfos, h.remoteInfos)
			}
		case GetUseCaseInfos:
			if nil != h.remoteInfos {
				frontend.sendUseCaseInfo(GetUseCaseInfos, h.useCaseInfos)
			}
		case GetAllData:
			sendData(h, data.SKI, data.Text)
		case SetConsumptionLimit:
			var limit = data.Limit

			h.consumptionLimits.IsActive = limit.IsActive
			h.consumptionLimits.Value = limit.Value
			h.consumptionLimits.Duration = limit.Duration * time.Second

			for _, remoteEntityScenario := range h.uclpc.RemoteEntitiesScenarios() {
				h.sendConsumptionLimit(remoteEntityScenario.Entity)
			}
		case SetProductionLimit:
			var limit = data.Limit

			h.productionLimits.IsActive = limit.IsActive
			h.productionLimits.Value = limit.Value
			h.productionLimits.Duration = limit.Duration * time.Second

			for _, remoteEntityScenario := range h.uclpp.RemoteEntitiesScenarios() {
				h.sendProductionLimit(remoteEntityScenario.Entity)
			}
		case SetConsumptionFailsafeValue:
			var limit = data.Value

			h.consumptionFailsafeLimits.Value = limit

			for _, remoteEntityScenario := range h.uclpc.RemoteEntitiesScenarios() {
				h.sendConsumptionFailsafeLimit(remoteEntityScenario.Entity)
			}
		case SetConsumptionFailsafeDuration:
			var limit = data.Value

			h.consumptionFailsafeLimits.Duration = time.Duration(limit) * time.Second

			for _, remoteEntityScenario := range h.uclpc.RemoteEntitiesScenarios() {
				h.sendConsumptionFailsafeDuration(remoteEntityScenario.Entity)
			}
		case SetProductionFailsafeValue:
			var limit = data.Value

			h.productionFailsafeLimits.Value = limit

			for _, remoteEntityScenario := range h.uclpp.RemoteEntitiesScenarios() {
				h.sendProductionFailsafeLimit(remoteEntityScenario.Entity)
			}
		case SetProductionFailsafeDuration:
			var limit = data.Value

			h.productionFailsafeLimits.Duration = time.Duration(limit) * time.Second

			for _, remoteEntityScenario := range h.uclpp.RemoteEntitiesScenarios() {
				h.sendProductionFailsafeDuration(remoteEntityScenario.Entity)
			}
			// TODO
			// case StopConsumptionHeartbeat:
			// 	h.uclpc.StopHeartbeat()
			// case StartConsumptionHeartbeat:
			// 	h.uclpc.StartHeartbeat()
		}

		frontend.sendNotification("", Acknowledge, "")
	}
}
