# ControlBox

ControlBox is a sample EEBUS GridGuard implementation that implements these EEBUS use cases:

- EnergyGuard Limitation of Power Consumption (LPC)
- EnergyGuard Limitation of Power Production (LPP)
- MonitoringAppliance Monitoring of Power Consumption (MPC)
- MonitoringAppliance Monitoring of Grid Connection Point (MGCP)

Forked from [vollautomat's eebus-go repository](https://github.com/vollautomat/eebus-go), based on [enbility's eebus-go implementation](https://github.com/enbility/eebus-go).

### Installation & Execution

#### ControlBox

Run ControlBox:
```
cd /path/to/controlbox
go run . 4712
2025-04-10 16:39:14 INFO  Local SKI: A46D9C217B8F335E921C4FAA087E615C9D2A73F0
```

Note the local SKI which is logged on ControlBox startup. Certificate and key are automatically created and saved to respective files.

#### evcc

As of evcc 0.301.0, EEBUS is enabled by default with certifacte/key being automatically created, dramatically simplifying setup.

Add ControlBox to the `evcc.yaml`:
```
hems:
  type: eebus
  ski: A46D9C217B8F335E921C4FAA087E615C9D2A73F0 # local SKI of the ControlBox
```

Restarting evcc will automatically connect evcc to the ControlBox.

#### ControlBox Frontend

Install dependencies:
```
cd /path/to/controlbox/frontend
npm install
```

Run web server:
```
npm run dev
```

Open ControlBox UI via web browser URI:
```
http://localhost:7081/
```


