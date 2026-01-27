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

#### Connect evcc

Generate EEBUS certificate and key for evcc via evcc CLI…
```
evcc eebus-cert
```

... and add them to the `evcc.yaml` config file:
```
eebus:
  certificate:
    public: |
      -----BEGIN CERTIFICATE-----
      MIIB3TCCAYOgAwIBAgIUWZp7lZ9JcM8xE5cQ6+4JkF0yZVswCgYIKoZIzj0EAwIw
      RDELMAkGA1UEBhMCREUxEjAQBgNVBAoMCUVFQlVTIFRFU1QxHTAbBgNVBAMMFEVF
      QlVTIERldmljZSBDZXJ0MB4XDTI2MDEyMDAwMDAwMFoXDTM2MDEyMDAwMDAwMFow
      RDELMAkGA1UEBhMCREUxEjAQBgNVBAoMCUVFQlVTIFRFU1QxHTAbBgNVBAMMFEVF
      QlVTIERldmljZSBDZXJ0MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE9G5xQK+S
      M6a9uY+qF8bB9n3GJ3Xh7V1Z5zM2Y3m1K1j9m1W5l0yBzYVZbF3gY7Cw7b7F1P0E
      Yt7D6kUq8aNTMFEwHQYDVR0OBBYEFN5C8sKx8c7K8Zx9+5J7n8Z2cJ5LMB8GA1Ud
      IwQYMBaAFN5C8sKx8c7K8Zx9+5J7n8Z2cJ5LMA8GA1UdEwEB/wQFMAMBAf8wCgYI
      KoZIzj0EAwIDSAAwRQIhAKF4Ewz5kD3qgC9Z7U8xZ5M1X8J6cJ3FZp6n3mMRAiB
      4YF8uQ5pJwYVZJ4n9GZPZJmM9H4k7Z0n9F5YVQ==
      -----END CERTIFICATE-----
    private: |
      -----BEGIN EC PRIVATE KEY-----
      MHcCAQEEIDP6JmP0kK6J7QFZ6NnZ4qFZQ1T+9n9C5k9Z8FZqz0ZBoAoGCCqGSM49
      AwEHoUQDQgAE9G5xQK+SM6a9uY+qF8bB9n3GJ3Xh7V1Z5zM2Y3m1K1j9m1W5l0yB
      zYVZbF3gY7Cw7b7F1P0EYt7D6kUq8Q==
      -----END EC PRIVATE KEY-----
```

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
