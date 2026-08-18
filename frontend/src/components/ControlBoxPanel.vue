<template>
  <h1>Control Box Simulator</h1>
  <div class="header-frame">
    <div class="header" v-if="'' < qrcode">
      <label class="qrcode-text">{{ qrcode }}</label>
      <qrcode-vue  class="qrcode" :value="qrcode" :size="130" level="H" render-as="svg" />
    </div>
  </div>
  <div v-if="'' == qrcode">
    <h3>not running</h3>
  </div>
  <div v-else>
    <div v-if="0 < remoteServices?.length" class="devices">
      <label class="device-select-label">Remote Device:</label>
      <VueSelect v-model="selectedSki" :options="optionServices"
        placeholder="Select a device" @option-selected="serviceSelected">
      </VueSelect>
    </div>
    <div v-else>
      <h3>No devices found</h3>
    </div>

    <div v-if="'' < selectedSki" class="devices">
      <label class="device-select-label">SKI:</label>
      <label class="device-select-label">{{ readableSKI( selectedSki ) }}</label>
      <label class="device-select-label">ID:</label>
      <label class="device-select-label">{{ deviceId }}</label>
      <label class="device-select-label">Device Brand:</label>
      <label class="device-select-label">{{ deviceBrand }}</label>
      <label class="device-select-label">Device Type:</label>
      <label class="device-select-label">{{ deviceType }}</label>
      <label class="device-select-label">Device Model:</label>
      <label class="device-select-label">{{ deviceModel }}</label>
      <label class="device-select-label">Serial Number:</label>
      <label class="device-select-label">{{ deviceSerial }}</label>
    </div>

    <div v-if="'' < selectedSki && !! remoteEntities" class="devices">
      <label class="device-select-label">Actors:</label>
      <VueSelect v-model="selectedActor" :options="optionActors"
        v-bind:placeholder="optionActors.length + (optionActors.length == 1 ? ' actor' : ' actors')">
      </VueSelect>
      <label class="device-select-label">Usecases:</label>
      <VueSelect :options="optionUsecases"
        v-bind:placeholder="! selectedActor ? '' : (optionUsecases.length + (optionUsecases.length == 1 ? ' usecase' : ' usecases'))">
      </VueSelect>
      <label class="device-select-label">Entities:</label>
      <VueSelect v-model="selectedEntity" :options="optionEntities"
        v-bind:placeholder="optionEntities.length + (optionEntities.length == 1 ? ' entity' : ' entities')">
      </VueSelect>
      <label class="device-select-label">Features:</label>
      <VueSelect :options="optionFeatures"
        v-bind:placeholder="! selectedEntity ? '' : (optionFeatures.length + (optionFeatures.length == 1 ? ' feature' : ' features'))">
      </VueSelect>
    </div>
    <div class="usecases">
      <div v-if="'' < selectedSki && !!selectedLs && !!selectedLs['LPC'] && existsUC('limitationOfPowerConsumption')">
        <h3>Consumption Limit</h3>
        <div class="form-line3">
          <label>Active:</label>
          <input type="checkbox" v-model="selectedLs['LPC'].IsActive" @input="onLPCUserChanged" />

          <label>Dimmed Value [W]:</label>
          <input type="number" v-model="selectedLs['LPC'].Value" @input="onLPCUserChanged" />
          <button class="three-lines" type="button" @click="setConsumptionLimit">Set</button>

          <label>Dimmed Duration [s]:</label>
          <input type="number" v-model="selectedLs['LPC'].Duration" @input="onLPCUserChanged" />

          <label>Failsafe Value [W]:</label>
          <input type="number" v-model="selectedLs['LPC'].FSValue" @input="onLPCUserChanged" />
          <button type="button" @click="setConsumptionFailsafeLimit">Set</button>

          <label>Failsafe Duration [s]:</label>
          <input type="number" v-model="selectedLs['LPC'].FSDuration" @input="onLPCUserChanged" />
          <button type="button" @click="setConsumptionFailsafeDuration">Set</button>

          <label>Nominal Maximum [W]:</label>
          <input type="number" v-model="consumptionNominalMax[selectedSki]" />
          <div></div>

          <label>Received Heartbeat:</label>
          <span v-bind:class = "(consumptionHeartbeat)?'pulse heartbeat':'pulse'">&#9673;</span>
          <!-- <button type="button" @click="toggleConsumptionHeartbeat">{{ consumptionHeartbeatEnabled ? 'Stop' : 'Start' }}</button> -->
          <div></div>
        </div>
      </div>

      <div v-if="'' < selectedSki && !!selectedLs && !!selectedLs['LPP'] && existsUC('limitationOfPowerProduction')">
        <h3>Production Limit</h3>
        <div class="form-line3">
          <label>Active:</label>
          <input type="checkbox" v-model="selectedLs['LPP'].IsActive" @input="onLPPUserChanged" />

          <label>Dimmed Value [W]:</label>
          <input type="number" v-model="selectedLs['LPP'].Value" @input="onLPPUserChanged" />
          <button class="three-lines" type="button" @click="setProductionLimit">Set</button>
          
          <label>Dimmed Duration [s]:</label>
          <input type="number" v-model="selectedLs['LPP'].Duration" @input="onLPPUserChanged" />
          
          <label>Failsafe Value [W]:</label>
          <input type="number" v-model="selectedLs['LPP'].FSValue" @input="onLPPUserChanged" />
          <button type="button" @click="setProductionFailsafeLimit">Set</button>
          
          <label>Failsafe Duration [s]:</label>
          <input type="number" v-model="selectedLs['LPP'].FSDuration" @input="onLPPUserChanged" />
          <button type="button" @click="setProductionFailsafeDuration">Set</button>
          
          <label>Nominal Maximum [W]:</label>
          <input type="number" v-model="productionNominalMax[selectedSki]" />
          <div></div>

          <label>Received Heartbeat:</label>
          <span v-bind:class = "(productionHeartbeat)?'pulse heartbeat':'pulse'">&#9673;</span>
          <!-- <button type="button" @click="toggleProductionHeartbeat">{{ productionHeartbeatEnabled ? 'Stop' : 'Start' }}</button> -->
          <div></div>
        </div>
      </div>

      <div v-if="'' < selectedSki && !!selectedMs && !!selectedMs['MGCP'] && existsUC('monitoringOfGridConnectionPoint')">
        <h3>Monitoring Grid Connection Point</h3>
        <div class="form-line2">
          <label>Power Limitation Factor:</label>
          <label>{{ selectedMs['MGCP'].PowerLimitationFactor ?? 0 }} %</label>

          <label>Power:</label>
          <label>{{ formatted( selectedMs['MGCP'].Power ?? 0 ) }} W</label>

          <label>Energy FeedIn:</label>
          <label>{{ formatted( selectedMs['MGCP'].EnergyFeedIn ?? 0 ) }} Wh</label>

          <label>Energy Consumed:</label>
          <label>{{ formatted( selectedMs['MGCP'].EnergyConsumed ?? 0 ) }} Wh</label>

          <label>Currents per Phase:</label>
          <label>{{ ! selectedMs['MGCP'].CurrentPerPhase ? '0' : selectedMs['MGCP'].CurrentPerPhase[0] }} A,
                 {{ ! selectedMs['MGCP'].CurrentPerPhase ? '0' : selectedMs['MGCP'].CurrentPerPhase[1] }} A,
                 {{ ! selectedMs['MGCP'].CurrentPerPhase ? '0' : selectedMs['MGCP'].CurrentPerPhase[2] }} A
          </label>

          <label>Voltages per Phase:</label>
          <label>{{ ! selectedMs['MGCP'].VoltagePerPhase ? '0' : selectedMs['MGCP'].VoltagePerPhase[0] }} V,
                 {{ ! selectedMs['MGCP'].VoltagePerPhase ? '0' : selectedMs['MGCP'].VoltagePerPhase[1] }} V,
                 {{ ! selectedMs['MGCP'].VoltagePerPhase ? '0' : selectedMs['MGCP'].VoltagePerPhase[2] }} V
          </label>

          <label>Frequency:</label>
          <label>{{ selectedMs['MGCP'].Frequency ?? 0 }} Hz</label>
        </div>
      </div>

      <div v-if="'' < selectedSki && !!selectedMs && !!selectedMs['MPC'] && existsUC('monitoringOfPowerConsumption')">
        <h3>Monitoring Power Consumption</h3>
        <div class="form-line2">
          <label>Power:</label>
          <label>{{ selectedMs['MPC'].Power ?? 0 }} W</label>

          <label>Power per Phase:</label>
          <label>{{ ! selectedMs['MPC'].PowerPerPhase ? '0' : selectedMs['MPC'].PowerPerPhase[0] }} W,
                 {{ ! selectedMs['MPC'].PowerPerPhase ? '0' : selectedMs['MPC'].PowerPerPhase[1] }} W,
                 {{ ! selectedMs['MPC'].PowerPerPhase ? '0' : selectedMs['MPC'].PowerPerPhase[2] }} W
          </label>

          <label>Energy FeedIn:</label>
          <label>{{ selectedMs['MPC'].EnergyFeedIn ?? 0 }} Wh</label>

          <label>Energy Consumed:</label>
          <label>{{ selectedMs['MPC'].EnergyConsumed ?? 0 }} Wh</label>

          <label>Currents per Phase:</label>
          <label>{{ ! selectedMs['MPC'].CurrentPerPhase ? '0' : selectedMs['MPC'].CurrentPerPhase[0] }} A,
                 {{ ! selectedMs['MPC'].CurrentPerPhase ? '0' : selectedMs['MPC'].CurrentPerPhase[1] }} A,
                 {{ ! selectedMs['MPC'].CurrentPerPhase ? '0' : selectedMs['MPC'].CurrentPerPhase[2] }} A
          </label>

          <label>Voltages per Phase:</label>
          <label>{{ ! selectedMs['MPC'].VoltagePerPhase ? '0' : selectedMs['MPC'].VoltagePerPhase[0] }} V,
                 {{ ! selectedMs['MPC'].VoltagePerPhase ? '0' : selectedMs['MPC'].VoltagePerPhase[1] }} V,
                 {{ ! selectedMs['MPC'].VoltagePerPhase ? '0' : selectedMs['MPC'].VoltagePerPhase[2] }} V
          </label>

          <label>Frequency:</label>
          <label>{{ selectedMs['MPC'].Frequency ?? 0 }} Hz</label>
        </div>
      </div>

    </div>
  </div>
</template>

<script lang="ts">
  import { Component, Vue, toNative } from 'vue-facing-decorator'
  import QrcodeVue from 'qrcode.vue'
  import VueSelect from 'vue3-select-component'

  enum MessageType {
    Text                           = 0,
    QRCode                         = 1,
    Acknowledge                    = 2,
    ServiceListChanged             = 3,
    GetServiceList                 = 4,
    SelectService                  = 5,
    GetEntityInfos                 = 6,
    GetUseCaseInfos                = 7,
    GetAllData                     = 8,
    SetConsumptionLimit            = 9,
    GetConsumptionLimit            = 10,
    SetProductionLimit             = 11,
    GetProductionLimit             = 12,
    SetConsumptionFailsafeValue    = 13,
    GetConsumptionFailsafeValue    = 14,
    SetConsumptionFailsafeDuration = 15,
    GetConsumptionFailsafeDuration = 16,
    SetProductionFailsafeValue     = 17,
    GetProductionFailsafeValue     = 18,
    SetProductionFailsafeDuration  = 19,
    GetProductionFailsafeDuration  = 20,
    GetConsumptionNominalMax       = 21,
    GetProductionNominalMax        = 22,
    GetConsumptionHeartbeat        = 23,
    StopConsumptionHeartbeat       = 24,
    StartConsumptionHeartbeat      = 25,
    GetProductionHeartbeat         = 26,
    StopProductionHeartbeat        = 27,
    StartProductionHeartbeat       = 28,
  	GetPowerLimitationFactor       = 29,
  	GetPower                       = 30,
    GetPowerPerPhase               = 31,
	  GetEnergyFeedIn                = 32,
	  GetEnergyConsumed              = 33,
	  GetCurrentPerPhase             = 34,
	  GetVoltagePerPhase             = 35,
	  GetFrequency                   = 36
}

  interface Limits {
    IsActive:   boolean,
	  Value:      number,
	  Duration:   number,
	  FSValue:    number,
	  FSDuration: number
  }

  interface Monitorings {
    PowerLimitationFactor: number,
    Power: number,
    PowerPerPhase: number[],
    EnergyFeedIn: number,
    EnergyConsumed: number,
    CurrentPerPhase: number[],
    VoltagePerPhase: number[],
    Frequency: number
  }

  interface RemoteService {
    name:       string,
	  ski:        string,
	  shipID:     string,
	  brand:      string,
	  type:       string,
	  model:      string,
	  serial:     string,
	  categories: number[]
  }

  interface EntityInfo {
    Address:  string,
    Name:     string,
    SKI:      string,
    Type:     string,
    Features: string[]
  }

  interface UseCaseInfo {
    Actor: string,
    Names: string[]
  }

  type UseCaseInfos = {[key:string]:UseCaseInfo[]}

  interface Message {
    SKI:           string,
    Type:          MessageType,
    Text?:         string,
    Limit?:        Limits,
    Value?:        number,
    Values?:       number[],
    ServiceList?:  RemoteService[],
    EntityInfos?:  EntityInfo[],
    UseCaseInfos?: UseCaseInfos
    UseCase?:      string
  }

  type UCLimits       = {[key:string]:Limits};
  type LimitData      = {[key:string]:UCLimits};
  type UCMonitorings  = {[key:string]:Monitorings};
  type MonitoringData = {[key:string]:UCMonitorings};

  @Component({
    components: {
      QrcodeVue,
      VueSelect
    }
  })
  export class ControlBoxPanel extends Vue {
    public qrcode = "";

    public limits: LimitData = {};
    public monitorings: MonitoringData = {};

    public remoteServices: RemoteService[] = [];
    public useCaseInfos: UseCaseInfos = {};
    public selectedActor: UseCaseInfo | undefined = undefined
    public remoteEntities: EntityInfo[] = [];
    public selectedSki = "";
    public selectedEntity: EntityInfo | undefined = undefined;

    private lpcUserChanged = false;
    private lppUserChanged = false;

    public existsUC( uc: string ): boolean {
      var exists = false;
      
      this.useCaseInfos[this.selectedSki]?.forEach( uci => {
        exists ||= uci.Names.includes( uc );
      } );

      return exists;
    }

    public get selectedLs() {
      return this.limits[this.selectedSki];
    }

    public get selectedMs() {
      return this.monitorings[this.selectedSki];
    }

    public get optionServices() {
      var options:any[] = [];
      this.remoteServices.forEach(item => { options.push({
          label: item.brand + " " + item.model + ("" < item.serial ? (", SN-" + item.serial) : ""),
          value: item.ski
        });        
      });
      return options;
    }

    public get optionActors() {
      var options:any[] = [];

      if ( "" < this.selectedSki && !! this.useCaseInfos[this.selectedSki]) {
        this.useCaseInfos[this.selectedSki].forEach(item => { options.push( {
            label: item.Actor,
            value: item
          });        
        });
      }
      return options;
    }

    public get optionUsecases() {
      var options:any[] = [];

      if ( "" < this.selectedSki && !! this.selectedActor ) {
        this.selectedActor.Names.forEach( item => { options.push( {
            label: item,
            value: item
          });        
        });
      }
      return options;
    }

    public get selectedEntities(): EntityInfo[] {
      if ( "" < this.selectedSki && !! this.remoteEntities ) {
        return this.remoteEntities.filter( re => re.SKI == this.selectedSki );
      }
      else {
        return [];
      }
    }

    public get optionEntities() {
      var options:any[] = [];

      if ( "" < this.selectedSki && !! this.selectedEntities ) {
        this.selectedEntities.forEach(item => { options.push({
            label: item.Name,
            value: item
          });        
        });
      }
      return options;
    }

    public get optionFeatures() {
      var options:any[] = [];

      if ( "" < this.selectedSki && !! this.selectedEntity ) {
        this.selectedEntity.Features.forEach(item => { options.push({
            label: item,
            value: item
          });        
        });
      }
      return options;
    }

    public get deviceId() {
      if ( "" < this.selectedSki ) {
        var remoteService = this.remoteServices.find( rs => rs.ski == this.selectedSki );
        return remoteService?.shipID ?? "";
      }
      else {
        return "";
      }
    }

    public get deviceBrand() {
      if ( "" < this.selectedSki ) {
        var remoteService = this.remoteServices.find( rs => rs.ski == this.selectedSki );
        return remoteService?.brand ?? "";
      }
      else {
        return "";
      }
    }

    public get deviceType() {
      if ( "" < this.selectedSki ) {
        var remoteService = this.remoteServices.find( rs => rs.ski == this.selectedSki );
        return remoteService?.type ?? "";
      }
      else {
        return "";
      }
    }

    public get deviceModel() {
      if ( "" < this.selectedSki ) {
        var remoteService = this.remoteServices.find( rs => rs.ski == this.selectedSki );
        return remoteService?.model ?? "";
      }
      else {
        return "";
      }
    }

    public get deviceSerial() {
      if ( "" < this.selectedSki ) {
        var remoteService = this.remoteServices.find( rs => rs.ski == this.selectedSki );
        return remoteService?.serial ?? "";
      }
      else {
        return "";
      }
    }

    public consumptionNominalMax: {[key: string]: number} = {};
    public productionNominalMax:  {[key: string]: number} = {};

    public consumptionHeartbeat:        boolean = false;
    public consumptionHeartbeatEnabled: boolean = true;
    public productionHeartbeat:         boolean = false;
    public productionHeartbeatEnabled:  boolean = true;

    private socket: WebSocket | undefined;
  
    mounted() {
      this.socket = new WebSocket( "ws://" + window.location.hostname + ":7080/ws" );
      console.log( "Attempting Connection..." );

      this.socket.onopen = () => {
          console.log( "Successfully Connected" );
          this.sendNotification( MessageType.GetEntityInfos );
      };
      
      this.socket.onclose = event => {
          console.log( "Socket Closed Connection: ", event );
          this.sendText( "Client Closed!" );
          this.socket = undefined;
      };

      this.socket.onerror = error => {
          console.log( "Socket Error: ", error );
      };

      this.socket.onmessage = event => {
        //console.log( "Socket message: ", event.data );
        var message: Message = JSON.parse( event.data );
        switch ( message.Type ) {
          case MessageType.QRCode: {
            this.qrcode = message.Text as string;
            console.log( "SHIPID: ", this.qrcode );
            break;
          }
          case MessageType.ServiceListChanged: {
            this.sendNotification( MessageType.GetServiceList );
            break;
          }
          case MessageType.SelectService: {
            if ( this.selectedSki == "" )
              this.selectedSki = message.Text ?? "";
            break;
          }
          case MessageType.GetServiceList: {
            this.remoteServices = message.ServiceList!;
            break;
          }
          case MessageType.GetEntityInfos: {
            this.syncEntities( message.EntityInfos );
            break;
          }
          case MessageType.GetUseCaseInfos: {
            this.syncUseCases( message.UseCaseInfos );
            break;
          }
          case MessageType.GetConsumptionLimit: {
            if ( ! this.lpcUserChanged ) {
              this.updateDeviceData( message.UseCase! );
              this.limits[message.SKI][message.UseCase!].IsActive = message.Limit?.IsActive ?? false;
              this.limits[message.SKI][message.UseCase!].Value    = message.Limit?.Value ?? 0;
              this.limits[message.SKI][message.UseCase!].Duration = message.Limit?.Duration ?? 0;
            }
            break;
          }
          case MessageType.GetProductionLimit: {
            if ( ! this.lppUserChanged ) {
              this.updateDeviceData( message.UseCase! );
              this.limits[message.SKI][message.UseCase!].IsActive = message.Limit?.IsActive ?? false;
              this.limits[message.SKI][message.UseCase!].Value    = message.Limit?.Value ?? 0;
              this.limits[message.SKI][message.UseCase!].Duration = message.Limit?.Duration ?? 0;
            }
            break;
          }
          case MessageType.GetConsumptionFailsafeValue: {
            if ( ! this.lpcUserChanged ) {
              this.updateDeviceData( message.UseCase! );
              this.limits[message.SKI][message.UseCase!].FSValue = message.Value ?? 0;
            }
            break;
          }
          case MessageType.GetProductionFailsafeValue: {
            if ( ! this.lppUserChanged ) {
              this.updateDeviceData( message.UseCase! );
              this.limits[message.SKI][message.UseCase!].FSValue = message.Value ?? 0;
            }
            break;
          }
          case MessageType.GetConsumptionFailsafeDuration: {
            if ( ! this.lpcUserChanged ) {
              this.updateDeviceData( message.UseCase! );
              this.limits[message.SKI][message.UseCase!].FSDuration = message.Value ?? 0;
            }
            break;
          }
          case MessageType.GetProductionFailsafeDuration: {
            if ( ! this.lppUserChanged ) {
              this.updateDeviceData( message.UseCase! );
              this.limits[message.SKI][message.UseCase!].FSDuration = message.Value ?? 0;
            }
            break;
          }
          case MessageType.GetConsumptionNominalMax: {
            this.consumptionNominalMax[message.SKI] = message.Value ?? 0;
            break;
          }
          case MessageType.GetProductionNominalMax: {
            this.productionNominalMax[message.SKI] = message.Value ?? 0;
            break;
          }
          case MessageType.GetConsumptionHeartbeat: {
            this.updateDeviceData( message.UseCase! );
            this.sendNotification( MessageType.GetAllData, message.UseCase );
            this.consumptionHeartbeat = true;
            setTimeout( () => this.consumptionHeartbeat = false, 1000 );
            break;
          }
          case MessageType.GetProductionHeartbeat: {
            this.updateDeviceData( message.UseCase! );
            this.sendNotification( MessageType.GetAllData, message.UseCase );
            this.productionHeartbeat = true;
            setTimeout( () => this.productionHeartbeat = false, 1000 );
            break;
          }
          case MessageType.GetPowerLimitationFactor: {
            this.updateDeviceData( message.UseCase! );
            this.monitorings[message.SKI][message.UseCase!].PowerLimitationFactor = message.Value ?? 0;
            break;
          }
	        case MessageType.GetPower: {
            this.updateDeviceData( message.UseCase! );
            this.monitorings[message.SKI][message.UseCase!].Power = message.Value ?? 0;
            break;
          }
        	case MessageType.GetPowerPerPhase: {
            this.updateDeviceData( message.UseCase! );
            this.monitorings[message.SKI][message.UseCase!].PowerPerPhase = message.Values ?? [0, 0, 0];
            break;
          }
        	case MessageType.GetEnergyFeedIn: {
            this.updateDeviceData( message.UseCase! );
            this.monitorings[message.SKI][message.UseCase!].EnergyFeedIn = message.Value ?? 0;
            break;
          }
        	case MessageType.GetEnergyConsumed: {
            this.updateDeviceData( message.UseCase! );
            this.monitorings[message.SKI][message.UseCase!].EnergyConsumed = message.Value ?? 0;
            break;
          }
        	case MessageType.GetCurrentPerPhase: {
            this.updateDeviceData( message.UseCase! );
            this.monitorings[message.SKI][message.UseCase!].CurrentPerPhase = message.Values ?? [0, 0, 0];
            break;
          }
        	case MessageType.GetVoltagePerPhase: {
            this.updateDeviceData( message.UseCase! );
            this.monitorings[message.SKI][message.UseCase!].VoltagePerPhase = message.Values ?? [0, 0, 0];
            break;
          }
        	case MessageType.GetFrequency: {
            this.updateDeviceData( message.UseCase! );
            this.monitorings[message.SKI][message.UseCase!].Frequency = message.Value ?? 0;
            break;
          }
        }   
      }
    }

    private updateDeviceData( useCase: string ) {
      if ( useCase == "LPC" || useCase == "LPP" ) {
        if ( ! this.limits[this.selectedSki] )
          this.limits[this.selectedSki] = {};
        if ( ! this.limits[this.selectedSki][useCase] )
          this.limits[this.selectedSki][useCase] = {} as Limits;
      }
      else if ( useCase == "MGCP" || useCase == "MPC" ) {
        if ( ! this.monitorings[this.selectedSki] )
          this.monitorings[this.selectedSki] = {};
        if ( ! this.monitorings[this.selectedSki][useCase] )
          this.monitorings[this.selectedSki][useCase] = {} as Monitorings;
      }
    }

    private syncEntities( entities: EntityInfo[] | undefined ) {
      if ( undefined == entities || 0 == entities.length ) {
        this.remoteEntities = [];
        return;
      }

      [...this.remoteEntities].forEach( re => {
        var indx = entities.findIndex( e => e.Address == re.Address );
        if ( -1 == indx ) {
          var indx2 = this.remoteEntities!.findIndex( e => e.Address == re.Address );
          this.remoteEntities = this.remoteEntities.splice( indx2, 1 );
        }
      } )

      entities.forEach( e => {
        var indx = this.remoteEntities.findIndex( re => re.Address == e.Address );
        if ( -1 == indx ) {
          this.remoteEntities.push( e );
        }
        else {
          this.remoteEntities[indx].Features = e.Features;
        }
      } );
    }

    private syncUseCases( useCaseInfos: UseCaseInfos | undefined ) {
      Object.keys( useCaseInfos! ).forEach( ski => {
        var infosOld = this.useCaseInfos[ski] ?? [];
        var infosNew = useCaseInfos![ski];

        [...infosOld].forEach( (infoOld, index) => {
            if ( undefined == infosNew.find( infoNew => infoNew.Actor == infoOld.Actor ) ) {
              infosOld = infosOld.splice( index, 1 );
            }
        } );

        infosNew.forEach( infoNew => {
          var infoOld = infosOld.find( infoOld => infoOld.Actor == infoNew.Actor )
          if ( undefined == infoOld )
            infosOld.push( infoNew );
          else
            infoOld.Names = infoNew.Names;    
        } );
        
        this.useCaseInfos[ski] = infosOld;
      } );
    }

    public serviceSelected() {
      this.selectedEntity = undefined;
      this.sendNotification( MessageType.SelectService, this.selectedSki );
    }

    private sendNotification( type: MessageType, param: string = "" ) {
      let command: Message = {
        SKI:  this.selectedSki,
        Type: type,
        Text: param
      };

      this.socket!.send( JSON.stringify( command ) );
    }

    private sendText( text: string ) {
      let command: Message = {
        SKI:  this.selectedSki,
        Type: MessageType.Text,
        Text: text,
      };

      this.socket!.send( JSON.stringify( command ) );
    }

    private sendLimits( type: MessageType, value: Limits ) {
      let command: Message = {
        SKI:   this.selectedSki,
        Type:  type,
        Limit: value
      };

      this.socket!.send( JSON.stringify( command ) );
    }

    private sendValue( type: MessageType, value: number ) {
      let command: Message = {
        SKI:   this.selectedSki,
        Type:  type,
        Value: value
      };

      this.socket!.send( JSON.stringify( command ) );
    }

    public onLPCUserChanged() {
      this.lpcUserChanged = true;
    }

    public onLPPUserChanged() {
      this.lppUserChanged = true;
    }

    public setConsumptionLimit() {
      if ( ! this.socket )
        return;
      
      this.sendLimits( MessageType.SetConsumptionLimit, this.limits[this.selectedSki]['LPC'] );
      this.lpcUserChanged = false;
    }

    public setProductionLimit() {
      if ( ! this.socket )
        return;
      
      this.sendLimits( MessageType.SetProductionLimit, this.limits[this.selectedSki]['LPP'] );
      this.lppUserChanged = false;
    }

    public setConsumptionFailsafeLimit() {
      if ( ! this.socket )
        return;
      
      this.sendValue( MessageType.SetConsumptionFailsafeValue, this.limits[this.selectedSki]['LPC'].FSValue );
      this.lpcUserChanged = false;
    }

    public setConsumptionFailsafeDuration() {
      if ( ! this.socket )
        return;
      
      this.sendValue( MessageType.SetConsumptionFailsafeDuration, this.limits[this.selectedSki]['LPC'].FSDuration );
      this.lpcUserChanged = false;
    }

    public setProductionFailsafeLimit() {
      if ( ! this.socket )
        return;
      
      this.sendValue( MessageType.SetProductionFailsafeValue, this.limits[this.selectedSki]['LPP'].FSValue );
      this.lppUserChanged = false;
    }

    public setProductionFailsafeDuration() {
      if ( ! this.socket )
        return;
      
      this.sendValue( MessageType.SetProductionFailsafeDuration, this.limits[this.selectedSki]['LPP'].FSDuration );
      this.lppUserChanged = false;
    }

    public toggleConsumptionHeartbeat() {
      if ( ! this.socket )
        return;

      if ( this.consumptionHeartbeatEnabled )
        this.sendValue( MessageType.StopConsumptionHeartbeat, 0 );
      else
        this.sendValue( MessageType.StartConsumptionHeartbeat, 0 );

      this.consumptionHeartbeatEnabled = ! this.consumptionHeartbeatEnabled;
    }

    public toggleProductionHeartbeat() {
      if ( ! this.socket )
        return;

      if ( this.productionHeartbeatEnabled )
        this.sendValue( MessageType.StopProductionHeartbeat, 0 );
      else
        this.sendValue( MessageType.StartProductionHeartbeat, 0 );

      this.productionHeartbeatEnabled = ! this.productionHeartbeatEnabled;
    }

    public readableSKI( ski: string ): string {
			if ( 40 < ski.length )
				return ski;

			var parts: string[] = [];
			for ( var i = 0; i < 10; i++ )
				parts[i] = ski.substring( 4*i, 4*i +4 );

			return parts.join( ' ' );
		}

    public formatted( val: number ): string {
			return new Intl.NumberFormat( 'en-US' ).format( val );
		}
  }

  export default toNative( ControlBoxPanel )
</script>

<style scoped>
  .read-the-docs {
    color: #888;
  }
  .header-frame {
    display: inline-table;
    margin-bottom: 10px;
  }
  .header {
    display: grid;
    grid-template-columns: 75fr 25fr;
    column-gap: 25px;
  }
  h1 {
    margin-top: 0.1em;
  }
  .qrcode-text {
    width: 370px;
    text-align: left;
    word-break: break-all;
  }
  .qrcode {
    width: 130px;
    height: 100%;
  }
  .usecases {
    display: grid;
    grid-template-columns: 50fr 50fr;
    column-gap: 25px;
  }
  .three-lines {
    grid-column-start: 3;
    grid-row-start: 1;
    grid-row-end: 4;
  }
  .form-line3 {
    display: grid;
    grid-template-columns: 50fr 30fr 20fr;
    column-gap: 10px;
  }
  .form-line3 label {
    align-content: center;
    text-align: left;
  }
  .form-line3 input {
    font-size: initial;
    width: 100px;
    align-self: center;
  }
  .form-line3 button {
    line-height: 5px;
    height: 100%;
  }
  
  .form-line2 {
    display: grid;
    grid-template-columns: 50fr 50fr;
    column-gap: 10px;
  }
  .form-line2 label {
    align-content: center;
    text-align: left;
  }

  .pulse {
    font-size: 25px;
  }

  .heartbeat {
    animation-name: heartbeat;
    animation-duration: 1s;
    /* animation-iteration-count: infinite; */
  }

  @keyframes heartbeat {
    from {
      color: rgb(0,255,0);
    }

    25% {
      color: rgb(0,255,0);
    }

    50% {
      color: rgb(0,127,0);
    }

    75% {
      color: rgb(0,63,0);
    }

    to {
      color: rgb(0,0,0);
    }
  }

  .devices {
    display: grid;
    grid-template-columns: 20fr 80fr;
    column-gap: 10px;
  }
  .device-select-label {
    text-align: left;
    line-height: 2.2em;
  }
</style>
