package main

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/enbility/eebus-go/api"
	"github.com/enbility/eebus-go/service"
	ucapi "github.com/enbility/eebus-go/usecases/api"
	"github.com/enbility/eebus-go/usecases/eg/lpc"
	"github.com/enbility/eebus-go/usecases/eg/lpp"
	"github.com/enbility/eebus-go/usecases/ma/mgcp"
	"github.com/enbility/eebus-go/usecases/ma/mpc"
	shipapi "github.com/enbility/ship-go/api"
	"github.com/enbility/ship-go/cert"
	spineapi "github.com/enbility/spine-go/api"
	"github.com/enbility/spine-go/model"
)

type failsafeLimits struct {
	Value    float64
	Duration time.Duration
}

type controlbox struct {
	myService *service.Service

	uclpc  ucapi.EgLPCInterface
	uclpp  ucapi.EgLPPInterface
	ucmgcp ucapi.MaMGCPInterface
	ucmpc  ucapi.MaMPCInterface

	isConnected map[string]bool

	remoteInfos  map[string]RemoteInfo
	useCaseInfos map[string][]UseCaseInfo

	consumptionLimits         ucapi.LoadLimit
	productionLimits          ucapi.LoadLimit
	consumptionFailsafeLimits failsafeLimits
	productionFailsafeLimits  failsafeLimits
	consumptionNominalMax     float64
	productionNominalMax      float64

	currentRemoteServices []shipapi.RemoteService

	mutex sync.Mutex
}

func loadOrCreateCertificate(dir string) (tls.Certificate, error) {
	crtPath := filepath.Join(dir, "cb.crt")
	keyPath := filepath.Join(dir, "cb.key")

	// If both files exist → load
	if _, err := os.Stat(crtPath); err == nil {
		if _, err := os.Stat(keyPath); err == nil {
			return tls.LoadX509KeyPair(crtPath, keyPath)
		}
	}

	// Otherwise create new certificate
	certTLS, err := cert.CreateCertificate("Demo", "Demo", "DE", "Demo-Unit-01")
	if err != nil {
		return tls.Certificate{}, err
	}

	// Write certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certTLS.Certificate[0],
	})
	if err := os.WriteFile(crtPath, certPEM, 0644); err != nil {
		return tls.Certificate{}, err
	}

	// Write private key
	privKey := certTLS.PrivateKey.(*ecdsa.PrivateKey)
	keyBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return tls.Certificate{}, err
	}

	return certTLS, nil
}

func (h *controlbox) run() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Println("Usage: controlbox <port> [cert-directory]")
		os.Exit(1)
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		usage()
		log.Fatal(err)
	}

	certDir := "."
	if len(os.Args) == 3 {
		certDir = os.Args[2]
	}

	certDir, err = filepath.Abs(certDir)
	if err != nil {
		log.Fatal(err)
	}

	certificate, err := loadOrCreateCertificate(certDir)
	if err != nil {
		log.Fatal(err)
	}

	vendorCode := "Demo"
	deviceBrand := "Demo"
	deviceModel := "ControlBox"
	serialNumber := "123456789"
	altIdentifier := "ControlBox Simulator SN-" + serialNumber

	h.isConnected = map[string]bool{}

	configuration, err := api.NewConfiguration(
		vendorCode, deviceBrand, deviceModel, serialNumber,
		[]shipapi.DeviceCategoryType{shipapi.DeviceCategoryTypeGridConnectionHub},
		model.DeviceTypeTypeElectricitySupplySystem,
		[]model.EntityTypeType{model.EntityTypeTypeGridGuard},
		port, certificate, time.Second*10)
	if err != nil {
		log.Fatal(err)
	}
	configuration.SetAlternateIdentifier(altIdentifier)

	h.myService = service.NewService(configuration, h)
	h.myService.SetLogging(h)

	if err = h.myService.Setup(); err != nil {
		fmt.Println(err)
		return
	}

	localEntity := h.myService.LocalDevice().EntityForType(model.EntityTypeTypeGridGuard)
	h.uclpc = lpc.NewLPC(localEntity, h.OnLPCEvent)
	h.myService.AddUseCase(h.uclpc)

	h.uclpp = lpp.NewLPP(localEntity, h.OnLPPEvent)
	h.myService.AddUseCase(h.uclpp)

	h.ucmgcp = mgcp.NewMGCP(localEntity, h.OnMGCPEvent)
	h.myService.AddUseCase(h.ucmgcp)

	h.ucmpc = mpc.NewMPC(localEntity, h.OnMPCEvent)
	h.myService.AddUseCase(h.ucmpc)

	h.remoteInfos = map[string]RemoteInfo{}
	h.useCaseInfos = map[string][]UseCaseInfo{}

	h.myService.Start()
}

// EEBUSServiceHandler

func (h *controlbox) RemoteSKIConnected(service api.ServiceInterface, ski string) {
	remoteSki = ski
	fmt.Println("RemoteSKIConnected: " + ski)
	h.isConnected[ski] = true

	frontend.sendText(SelectService, ski)
}

func (h *controlbox) RemoteSKIDisconnected(service api.ServiceInterface, ski string) {
	fmt.Println("RemoteSKIDisconnected: " + ski)
	h.isConnected[ski] = false

	frontend.sendNotification("", ServiceListChanged, "")
}

func (h *controlbox) VisibleRemoteServicesUpdated(service api.ServiceInterface, entries []shipapi.RemoteService) {
	fmt.Print("VisibleRemoteServicesUpdated, count: ")
	fmt.Println(len(entries))

	for _, element := range entries {
		fmt.Println("Remote SKI: " + element.Ski)
		service := h.myService.RemoteServiceForSKI(element.Ski)
		service.SetTrusted(true)
	}

	h.currentRemoteServices = entries

	frontend.sendNotification("", ServiceListChanged, "")
}

func (h *controlbox) ServiceShipIDUpdate(ski string, shipdID string) {
}

func (h *controlbox) ServicePairingDetailUpdate(ski string, detail *shipapi.ConnectionStateDetail) {
	if ski == remoteSki && detail.State() == shipapi.ConnectionStateRemoteDeniedTrust {
		fmt.Println("The remote service denied trust. Exiting.")
		h.myService.CancelPairingWithSKI(ski)
		h.myService.UnregisterRemoteSKI(ski)
		h.myService.Shutdown()
		os.Exit(1)
	}

	frontend.sendNotification("", ServiceListChanged, "")
}

func (h *controlbox) AllowWaitingForTrust(ski string) bool {
	//return ski == remoteSki
	fmt.Println("AllowWaitingForTrust: " + ski)
	return true
}

func (h *controlbox) updateEntityInfos(ski string, device spineapi.DeviceRemoteInterface, uc string) {
	info, exists := h.remoteInfos[ski]
	if !exists {
		indx := slices.IndexFunc(h.currentRemoteServices, func(v shipapi.RemoteService) bool { return v.Ski == ski })
		h.remoteInfos[ski] = RemoteInfo{
			Service:  h.currentRemoteServices[indx],
			Device:   device,
			UseCases: []string{uc},
		}
	} else {
		info.Device = device
		found := slices.Contains(info.UseCases, uc)
		if !found {
			info.UseCases = append(info.UseCases, uc)
			h.remoteInfos[ski] = info
		}
	}
}

func (h *controlbox) updateUseCaseInfos(ski string, device spineapi.DeviceRemoteInterface) {
	info := []UseCaseInfo{}

	for _, uc := range device.UseCases() {
		actor := string(*uc.Actor)
		names := []string{}
		for _, ucs := range uc.UseCaseSupport {
			names = append(names, string(*ucs.UseCaseName))
		}

		info = append(info, UseCaseInfo{
			Actor: actor,
			Names: names,
		})
	}

	h.useCaseInfos[ski] = info
}

// LPC Event Handler

func (h *controlbox) sendConsumptionLimit(entity spineapi.EntityRemoteInterface) {
	resultCB := func(msg model.ResultDataType) {
		if *msg.ErrorNumber == model.ErrorNumberTypeNoError {
			fmt.Println("Consumption limit accepted.")
		} else {
			fmt.Println("Consumption limit rejected. Code", *msg.ErrorNumber, "Description", *msg.Description)
		}
	}
	msgCounter, err := h.uclpc.WriteConsumptionLimit(entity, h.consumptionLimits, resultCB)
	if err != nil {
		fmt.Println("Failed to send consumption limit", err)
		return
	}
	fmt.Println("Sent consumption limit to", entity.Device().Ski(), "with msgCounter", msgCounter)
}

func (h *controlbox) sendConsumptionFailsafeLimit(entity spineapi.EntityRemoteInterface) {
	msgCounter, err := h.uclpc.WriteFailsafeConsumptionActivePowerLimit(entity, h.consumptionFailsafeLimits.Value)
	if err != nil {
		fmt.Println("Failed to send consumption failsafe limit", err)
		return
	}
	fmt.Println("Sent consumption failsafe limit to", entity.Device().Ski(), "with msgCounter", msgCounter)
}

func (h *controlbox) sendConsumptionFailsafeDuration(entity spineapi.EntityRemoteInterface) {
	msgCounter, err := h.uclpc.WriteFailsafeDurationMinimum(entity, h.consumptionFailsafeLimits.Duration)
	if err != nil {
		fmt.Println("Failed to send consumption failsafe duration", err)
		return
	}
	fmt.Println("Sent consumption failsafe duration to", entity.Device().Ski(), "with msgCounter", msgCounter)
}

func (h *controlbox) readConsumptionNominalMax(entity spineapi.EntityRemoteInterface) {
	nominal, err := h.uclpc.ConsumptionNominalMax(entity)

	if err != nil {
		fmt.Println("Failed to get consumption nominal max", err)
		return
	}

	frontend.sendValue(entity.Device().Ski(), GetConsumptionNominalMax, "LPC", nominal)
}

func (h *controlbox) OnLPCEvent(ski string, device spineapi.DeviceRemoteInterface, entity spineapi.EntityRemoteInterface, event api.EventType) {
	fmt.Println("--> LPC Event: " + string(event) + " from " + ski)
	connected, exists := h.isConnected[ski]
	if !exists || !connected {
		fmt.Println("--> but not connected")
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.updateEntityInfos(ski, device, "LPC")
	frontend.sendEntityInfo(GetEntityInfos, h.remoteInfos)
	h.updateUseCaseInfos(ski, device)
	frontend.sendUseCaseInfo(GetUseCaseInfos, h.useCaseInfos)

	switch event {
	case lpc.UseCaseSupportUpdate:
		readData(h, entity, []string{"LPC"})

	case lpc.DataUpdateLimit:
		if currentLimit, err := h.uclpc.ConsumptionLimit(entity); err == nil {
			if ski == remoteSki {
				fmt.Println("Event lpc.DataUpdateLimit", ski, currentLimit.Value)

				h.consumptionLimits = currentLimit

				if currentLimit.IsActive {
					fmt.Println("New consumption limit received: active,", currentLimit.Value, "W,", currentLimit.Duration)
				} else {
					fmt.Println("New consumption limit received: inactive,", currentLimit.Value, "W,", currentLimit.Duration)
				}
				frontend.sendLimit(ski, GetConsumptionLimit, "LPC", ucapi.LoadLimit{
					IsActive: currentLimit.IsActive,
					Duration: currentLimit.Duration / time.Second,
					Value:    currentLimit.Value})
			}
		}
	case lpc.DataUpdateFailsafeConsumptionActivePowerLimit:
		if limit, err := h.uclpc.FailsafeConsumptionActivePowerLimit(entity); err == nil {
			if ski == remoteSki {
				fmt.Println("Event lpc.DataUpdateFailsafeConsumptionActivePowerLimit", ski, limit)

				h.consumptionFailsafeLimits.Value = limit

				frontend.sendValue(ski, GetConsumptionFailsafeValue, "LPC", limit)
			}
		}
	case lpc.DataUpdateFailsafeDurationMinimum:
		if duration, err := h.uclpc.FailsafeDurationMinimum(entity); err == nil {
			if ski == remoteSki {
				fmt.Println("Event lpc.DataUpdateFailsafeDurationMinimum", ski, duration)

				h.consumptionFailsafeLimits.Duration = duration

				frontend.sendValue(ski, GetConsumptionFailsafeDuration, "LPC", float64(duration/time.Second))
			}
		}
		// TODO
	// case lpc.DataUpdateHeartbeat:
	// 	if ski == remoteSki {
	// 		h.readConsumptionNominalMax(entity)
	// 		frontend.sendNotification(ski, GetConsumptionHeartbeat, "LPC")
	// 	}
	default:
		return
	}
}

// LPP Event Handler

func (h *controlbox) sendProductionLimit(entity spineapi.EntityRemoteInterface) {
	resultCB := func(msg model.ResultDataType) {
		if *msg.ErrorNumber == model.ErrorNumberTypeNoError {
			fmt.Println("Production limit accepted.")
		} else {
			fmt.Println("Production limit rejected. Code", *msg.ErrorNumber, "Description", *msg.Description)
		}
	}
	msgCounter, err := h.uclpp.WriteProductionLimit(entity, h.productionLimits, resultCB)
	if err != nil {
		fmt.Println("Failed to send production limit", err)
		return
	}
	fmt.Println("Sent production limit to", entity.Device().Ski(), "with msgCounter", msgCounter)
}

func (h *controlbox) sendProductionFailsafeLimit(entity spineapi.EntityRemoteInterface) {
	msgCounter, err := h.uclpp.WriteFailsafeProductionActivePowerLimit(entity, h.productionFailsafeLimits.Value)
	if err != nil {
		fmt.Println("Failed to send production failsafe limit", err)
		return
	}
	fmt.Println("Sent production failsafe limit to", entity.Device().Ski(), "with msgCounter", msgCounter)
}

func (h *controlbox) sendProductionFailsafeDuration(entity spineapi.EntityRemoteInterface) {
	msgCounter, err := h.uclpp.WriteFailsafeDurationMinimum(entity, h.productionFailsafeLimits.Duration)
	if err != nil {
		fmt.Println("Failed to send production failsafe duration", err)
		return
	}
	fmt.Println("Sent production failsafe duration to", entity.Device().Ski(), "with msgCounter", msgCounter)
}

func (h *controlbox) readProductionNominalMax(entity spineapi.EntityRemoteInterface) {
	nominal, err := h.uclpp.ProductionNominalMax(entity)

	if err != nil {
		fmt.Println("Failed to get production nominal max", err)
		return
	}

	frontend.sendValue(entity.Device().Ski(), GetProductionNominalMax, "LPP", nominal)
}

func (h *controlbox) OnLPPEvent(ski string, device spineapi.DeviceRemoteInterface, entity spineapi.EntityRemoteInterface, event api.EventType) {
	fmt.Println("--> LPP Event: " + string(event) + " from " + ski)
	connected, exists := h.isConnected[ski]
	if !exists || !connected {
		fmt.Println("--> but not connected")
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.updateEntityInfos(ski, device, "LPP")
	frontend.sendEntityInfo(GetEntityInfos, h.remoteInfos)
	h.updateUseCaseInfos(ski, device)
	frontend.sendUseCaseInfo(GetUseCaseInfos, h.useCaseInfos)

	switch event {
	case lpp.UseCaseSupportUpdate:
		readData(h, entity, []string{"LPP"})

	case lpp.DataUpdateLimit:
		if currentLimit, err := h.uclpp.ProductionLimit(entity); err == nil {
			if ski == remoteSki {
				fmt.Println("Event lpp.DataUpdateLimit", ski, currentLimit.Value)

				h.productionLimits = currentLimit

				if currentLimit.IsActive {
					fmt.Println("New production limit received: active,", currentLimit.Value, "W,", currentLimit.Duration)
				} else {
					fmt.Println("New production limit received: inactive,", currentLimit.Value, "W,", currentLimit.Duration)
				}

				frontend.sendLimit(ski, GetProductionLimit, "LPP", ucapi.LoadLimit{
					IsActive: currentLimit.IsActive,
					Duration: currentLimit.Duration / time.Second,
					Value:    currentLimit.Value})
			}
		}
	case lpp.DataUpdateFailsafeProductionActivePowerLimit:
		if limit, err := h.uclpp.FailsafeProductionActivePowerLimit(entity); err == nil {
			if ski == remoteSki {
				fmt.Println("Event lpp.DataUpdateFailsafeProductionActivePowerLimit", ski, limit)

				h.productionFailsafeLimits.Value = limit

				frontend.sendValue(ski, GetProductionFailsafeValue, "LPP", limit)
			}
		}
	case lpp.DataUpdateFailsafeDurationMinimum:
		if duration, err := h.uclpp.FailsafeDurationMinimum(entity); err == nil {
			if ski == remoteSki {
				fmt.Println("Event lpp.DataUpdateFailsafeDurationMinimum", ski, duration)

				h.productionFailsafeLimits.Duration = duration

				frontend.sendValue(ski, GetProductionFailsafeDuration, "LPP", float64(duration/time.Second))
			}
		}
		// TODO
	// case lpp.DataUpdateHeartbeat:
	// 	if ski == remoteSki {
	// 		h.readProductionNominalMax(entity)
	// 		frontend.sendNotification(ski, GetProductionHeartbeat, "LPP")
	// 	}
	default:
		return
	}
}

func (h *controlbox) OnMGCPEvent(ski string, device spineapi.DeviceRemoteInterface, entity spineapi.EntityRemoteInterface, event api.EventType) {
	fmt.Println("--> MGCP Event: " + string(event) + " from " + ski)
	connected, exists := h.isConnected[ski]
	if !exists || !connected {
		fmt.Println("--> but not connected")
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.updateEntityInfos(ski, device, "MGCP")
	frontend.sendEntityInfo(GetEntityInfos, h.remoteInfos)
	h.updateUseCaseInfos(ski, device)
	frontend.sendUseCaseInfo(GetUseCaseInfos, h.useCaseInfos)

	switch event {
	case mgcp.UseCaseSupportUpdate:
		readData(h, entity, []string{"MGCP"})

	case mgcp.DataUpdatePowerLimitationFactor:
		if powerLimitFactor, err := h.ucmgcp.PowerLimitationFactor(entity); err == nil {
			frontend.sendValue(ski, GetPowerLimitationFactor, "MGCP", powerLimitFactor)
		}
	case mgcp.DataUpdatePower:
		if power, err := h.ucmgcp.Power(entity); err == nil {
			frontend.sendValue(ski, GetPower, "MGCP", power)
		}
	case mgcp.DataUpdateEnergyFeedIn:
		if energyFeedIn, err := h.ucmgcp.EnergyFeedIn(entity); err == nil {
			frontend.sendValue(ski, GetEnergyFeedIn, "MGCP", energyFeedIn)
		}
	case mgcp.DataUpdateEnergyConsumed:
		if energyConsumed, err := h.ucmgcp.EnergyConsumed(entity); err == nil {
			frontend.sendValue(ski, GetEnergyConsumed, "MGCP", energyConsumed)
		}
	case mgcp.DataUpdateCurrentPerPhase:
		if currentPerPhase, err := h.ucmgcp.CurrentPerPhase(entity); err == nil {
			frontend.sendValueArr(ski, GetCurrentPerPhase, "MGCP", currentPerPhase)
		}
	case mgcp.DataUpdateVoltagePerPhase:
		if voltagePerPhase, err := h.ucmgcp.VoltagePerPhase(entity); err == nil {
			frontend.sendValueArr(ski, GetVoltagePerPhase, "MGCP", voltagePerPhase)
		}
	case mgcp.DataUpdateFrequency:
		if frequency, err := h.ucmgcp.Frequency(entity); err == nil {
			frontend.sendValue(ski, GetFrequency, "MGCP", frequency)
		}
	}
}

func (h *controlbox) OnMPCEvent(ski string, device spineapi.DeviceRemoteInterface, entity spineapi.EntityRemoteInterface, event api.EventType) {
	fmt.Println("--> MPC Event: " + string(event) + " from " + ski)
	connected, exists := h.isConnected[ski]
	if !exists || !connected {
		fmt.Println("--> but not connected")
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.updateEntityInfos(ski, device, "MPC")
	frontend.sendEntityInfo(GetEntityInfos, h.remoteInfos)
	h.updateUseCaseInfos(ski, device)
	frontend.sendUseCaseInfo(GetUseCaseInfos, h.useCaseInfos)

	switch event {
	case mpc.UseCaseSupportUpdate:
		readData(h, entity, []string{"MPC"})

	case mpc.DataUpdatePower:
		if power, err := h.ucmpc.Power(entity); err == nil {
			frontend.sendValue(ski, GetPower, "MPC", power)
		}
	case mpc.DataUpdatePowerPerPhase:
		if powerPerPhase, err := h.ucmpc.PowerPerPhase(entity); err == nil {
			frontend.sendValueArr(ski, GetPowerPerPhase, "MPC", powerPerPhase)
		}
	case mpc.DataUpdateEnergyConsumed:
		if energyConsumed, err := h.ucmpc.EnergyConsumed(entity); err == nil {
			frontend.sendValue(ski, GetEnergyConsumed, "MPC", energyConsumed)
		}
	case mpc.DataUpdateEnergyProduced:
		if energyFeedIn, err := h.ucmpc.EnergyProduced(entity); err == nil {
			frontend.sendValue(ski, GetEnergyFeedIn, "MPC", energyFeedIn)
		}
	case mpc.DataUpdateCurrentsPerPhase:
		if currentPerPhase, err := h.ucmpc.CurrentPerPhase(entity); err == nil {
			frontend.sendValueArr(ski, GetCurrentPerPhase, "MPC", currentPerPhase)
		}
	case mpc.DataUpdateVoltagePerPhase:
		if voltagePerPhase, err := h.ucmpc.VoltagePerPhase(entity); err == nil {
			frontend.sendValueArr(ski, GetVoltagePerPhase, "MPC", voltagePerPhase)
		}
	case mpc.DataUpdateFrequency:
		if frequency, err := h.ucmpc.Frequency(entity); err == nil {
			frontend.sendValue(ski, GetFrequency, "MPC", frequency)
		}
	}
}
