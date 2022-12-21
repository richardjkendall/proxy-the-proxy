```
   ___                       _   _                                      
  / _ \_ __ _____  ___   _  | |_| |__   ___   _ __  _ __ _____  ___   _ 
 / /_)/ '__/ _ \ \/ / | | | | __| '_ \ / _ \ | '_ \| '__/ _ \ \/ / | | |
/ ___/| | | (_) >  <| |_| | | |_| | | |  __/ | |_) | | | (_) >  <| |_| |
\/    |_|  \___/_/\_\\__, |  \__|_| |_|\___| | .__/|_|  \___/_/\_\\__, |
                     |___/                   |_|                  |___/ 
```

## What is this?

Some corporate networks use 'automatic proxy detection' which is a protocol to allow a computer to auto-discover proxies to be used for accessing outside websites.  

Many command line tools don't play nicely with this and instead rely on proxies being set in environment variables or configration files.  This is fine if your environment is static, but when it is not, it can be pain to discover the correct address and set it manually.

So I created this tool, which acts a local proxy and which can use the WPAD protocol to discover which proxy server can be used.  It means you can set a local address as your HTTP and HTTPS proxy and then not worry about what's going on in the background.

## Features

* Support for HTTP and HTTPS (direct and via CONNECT)
* Javascript interpreter built-in for running PAC files
* Caching of proxy address details to speed up (PAC is only)
* Built-in management server to control the proxy
* Prometheus exporter for metrics

## How does it work

By default the proxy listens on port 8080 and the management server on 9001.

On startup it checks to see if a proxy can be discovered via the default outbound interface.  If it can then it configures itself to use the auto-discovery config file (the PAC file).  If no proxy can be discovered it configures itself for direct access.

Requests sent to the proxy port will then be proxied to their destination, via the intermediate proxy if one is discovered.

### Changing default ports

You can change the ports which the tool listens on using command line parameters.  

## Management server

The management server offers the following endpoints.

| Endpoint | Method | Purpose |
| --- | --- | --- |
|`/`| `GET` | Provides a status of the service
|`/metrics`| `GET` | Prometheus metrics endpoint
|`/refresh`| `GET` | Refresh the IP address and auto-detected proxy details

## TODO

* Auto-detect network interface changes
* Support proxy authentication
* Complete unit test coverage