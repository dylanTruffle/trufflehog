# Detector Profiling Report

This report shows detailed performance metrics for each detector during the scan.

## Summary

- **Total Detectors**: 355
- **Total Execution Time**: 7m17.456701s
- **Total Detector Calls**: 505
- **Average Time per Call**: 866.251ms

## Top 10 Slowest Detectors (by total time)

1. **JDBC**
   - Total Time: 2m30.138951s
   - Calls: 18
   - Average: 8.341053s
   - Min: 27.154ms
   - Max: 10.001039s

2. **Couchbase**
   - Total Time: 1m0.053961s
   - Calls: 1
   - Average: 1m0.053961s
   - Min: 1m0.053961s
   - Max: 1m0.053961s

3. **PrivateKey**
   - Total Time: 23.160692s
   - Calls: 31
   - Average: 747.119ms
   - Min: 109.962ms
   - Max: 5.508684s

4. **Postgres**
   - Total Time: 18.006033s
   - Calls: 3
   - Average: 6.002011s
   - Min: 4.341ms
   - Max: 9.000862s

5. **MongoDB**
   - Total Time: 12.009409s
   - Calls: 4
   - Average: 3.002352s
   - Min: 91µs
   - Max: 6.001882s

6. **Billomat**
   - Total Time: 10.065554s
   - Calls: 2
   - Average: 5.032777s
   - Min: 4.055643s
   - Max: 6.009911s

7. **PubNubSubscriptionKey**
   - Total Time: 10.001762s
   - Calls: 2
   - Average: 5.000881s
   - Min: 5.000816s
   - Max: 5.000946s

8. **AzureContainerRegistry**
   - Total Time: 6.743991s
   - Calls: 4
   - Average: 1.685998s
   - Min: 14.784ms
   - Max: 6.687123s

9. **Redis**
   - Total Time: 5.005351s
   - Calls: 1
   - Average: 5.005351s
   - Min: 5.005351s
   - Max: 5.005351s

10. **PubNubPublishKey**
   - Total Time: 5.000972s
   - Calls: 1
   - Average: 5.000972s
   - Min: 5.000972s
   - Max: 5.000972s

## Complete Detector Performance Table

| Detector | Calls | Total Time | Average | Min | Max |
|----------|-------|------------|---------|-----|-----|
| JDBC | 18 | 2m30.138951s | 8.341053s | 27.154ms | 10.001039s |
| Couchbase | 1 | 1m0.053961s | 1m0.053961s | 1m0.053961s | 1m0.053961s |
| PrivateKey | 31 | 23.160692s | 747.119ms | 109.962ms | 5.508684s |
| Postgres | 3 | 18.006033s | 6.002011s | 4.341ms | 9.000862s |
| MongoDB | 4 | 12.009409s | 3.002352s | 91µs | 6.001882s |
| Billomat | 2 | 10.065554s | 5.032777s | 4.055643s | 6.009911s |
| PubNubSubscriptionKey | 2 | 10.001762s | 5.000881s | 5.000816s | 5.000946s |
| AzureContainerRegistry | 4 | 6.743991s | 1.685998s | 14.784ms | 6.687123s |
| Redis | 1 | 5.005351s | 5.005351s | 5.005351s | 5.005351s |
| PubNubPublishKey | 1 | 5.000972s | 5.000972s | 5.000972s | 5.000972s |
| Holistic | 1 | 5.000564s | 5.000564s | 5.000564s | 5.000564s |
| Campayn | 1 | 5.000363s | 5.000363s | 5.000363s | 5.000363s |
| Twitch | 1 | 4.235929s | 4.235929s | 4.235929s | 4.235929s |
| FTP | 2 | 3.857386s | 1.928693s | 59.626ms | 3.79776s |
| SumoLogicKey | 1 | 3.342987s | 3.342987s | 3.342987s | 3.342987s |
| CaptainData | 2 | 2.618958s | 1.309479s | 1.254279s | 1.364679s |
| ScrapingBee | 1 | 2.583589s | 2.583589s | 2.583589s | 2.583589s |
| SendinBlueV2 | 1 | 2.565731s | 2.565731s | 2.565731s | 2.565731s |
| Gitlab | 8 | 2.345849s | 293.231ms | 164.326ms | 708.617ms |
| SquareApp | 2 | 2.307559s | 1.15378s | 627.351ms | 1.680209s |
| Dwolla | 1 | 2.276653s | 2.276653s | 2.276653s | 2.276653s |
| Mrticktock | 2 | 2.252349s | 1.126175s | 1.114504s | 1.137846s |
| Fibery | 1 | 2.163454s | 2.163454s | 2.163454s | 2.163454s |
| Azure | 5 | 1.824734s | 364.947ms | 46.978ms | 814.219ms |
| Fleetbase | 1 | 1.760474s | 1.760474s | 1.760474s | 1.760474s |
| FormBucket | 3 | 1.724432s | 574.811ms | 245.114ms | 745.535ms |
| Clearbit | 1 | 1.637079s | 1.637079s | 1.637079s | 1.637079s |
| Klaviyo | 1 | 1.436213s | 1.436213s | 1.436213s | 1.436213s |
| AmplitudeApiKey | 1 | 1.336495s | 1.336495s | 1.336495s | 1.336495s |
| CompanyHub | 1 | 1.293077s | 1.293077s | 1.293077s | 1.293077s |
| Dotmailer | 1 | 1.166973s | 1.166973s | 1.166973s | 1.166973s |
| Bitfinex | 1 | 1.115921s | 1.115921s | 1.115921s | 1.115921s |
| Deputy | 1 | 1.090348s | 1.090348s | 1.090348s | 1.090348s |
| ZipAPI | 1 | 1.064893s | 1.064893s | 1.064893s | 1.064893s |
| APITemplate | 1 | 1.059888s | 1.059888s | 1.059888s | 1.059888s |
| PosthogApp | 1 | 1.039632s | 1.039632s | 1.039632s | 1.039632s |
| AWS | 6 | 955.512ms | 159.252ms | 36.913ms | 677.223ms |
| CraftMyPDF | 1 | 950.859ms | 950.859ms | 950.859ms | 950.859ms |
| Codequiry | 1 | 893.601ms | 893.601ms | 893.601ms | 893.601ms |
| Finage | 1 | 856.854ms | 856.854ms | 856.854ms | 856.854ms |
| FileIO | 1 | 847.493ms | 847.493ms | 847.493ms | 847.493ms |
| BuddyNS | 1 | 833.96ms | 833.96ms | 833.96ms | 833.96ms |
| GetSandbox | 1 | 805.246ms | 805.246ms | 805.246ms | 805.246ms |
| URI | 34 | 790.795ms | 23.259ms | 132µs | 365.1ms |
| OpenVpn | 1 | 788.956ms | 788.956ms | 788.956ms | 788.956ms |
| CloudflareApiToken | 1 | 780.352ms | 780.352ms | 780.352ms | 780.352ms |
| Dynalist | 1 | 776.196ms | 776.196ms | 776.196ms | 776.196ms |
| Geocodio | 1 | 768.52ms | 768.52ms | 768.52ms | 768.52ms |
| Fmfw | 1 | 760.165ms | 760.165ms | 760.165ms | 760.165ms |
| Diggernaut | 1 | 750.8ms | 750.8ms | 750.8ms | 750.8ms |
| CurrentsAPI | 1 | 732.559ms | 732.559ms | 732.559ms | 732.559ms |
| EightxEight | 1 | 731.337ms | 731.337ms | 731.337ms | 731.337ms |
| Enablex | 1 | 705.339ms | 705.339ms | 705.339ms | 705.339ms |
| DigitalOceanV2 | 1 | 690.626ms | 690.626ms | 690.626ms | 690.626ms |
| Besttime | 2 | 690.607ms | 345.304ms | 335.819ms | 354.788ms |
| OneLogin | 1 | 683.802ms | 683.802ms | 683.802ms | 683.802ms |
| GraphCMS | 1 | 677.62ms | 677.62ms | 677.62ms | 677.62ms |
| FlightLabs | 1 | 658.33ms | 658.33ms | 658.33ms | 658.33ms |
| SatismeterProjectkey | 1 | 654.777ms | 654.777ms | 654.777ms | 654.777ms |
| AzureSearchAdminKey | 2 | 642.189ms | 321.094ms | 320.985ms | 321.204ms |
| ClustDoc | 1 | 630.371ms | 630.371ms | 630.371ms | 630.371ms |
| CurrencyScoop | 1 | 614.051ms | 614.051ms | 614.051ms | 614.051ms |
| CarbonInterface | 1 | 607.369ms | 607.369ms | 607.369ms | 607.369ms |
| Mailgun | 3 | 597.538ms | 199.179ms | 91.872ms | 390.485ms |
| Commodities | 1 | 590.662ms | 590.662ms | 590.662ms | 590.662ms |
| HelpCrunch | 1 | 583.703ms | 583.703ms | 583.703ms | 583.703ms |
| RazorPay | 1 | 582.603ms | 582.603ms | 582.603ms | 582.603ms |
| CexIO | 1 | 581.852ms | 581.852ms | 581.852ms | 581.852ms |
| Formcraft | 1 | 580.522ms | 580.522ms | 580.522ms | 580.522ms |
| Currencylayer | 1 | 568.096ms | 568.096ms | 568.096ms | 568.096ms |
| Bitmex | 1 | 558.034ms | 558.034ms | 558.034ms | 558.034ms |
| Collect2 | 1 | 549.438ms | 549.438ms | 549.438ms | 549.438ms |
| Replicate | 1 | 538.236ms | 538.236ms | 538.236ms | 538.236ms |
| AviationStack | 1 | 522.397ms | 522.397ms | 522.397ms | 522.397ms |
| Html2Pdf | 1 | 518.684ms | 518.684ms | 518.684ms | 518.684ms |
| ClickSendsms | 1 | 517.604ms | 517.604ms | 517.604ms | 517.604ms |
| Flutterwave | 1 | 512.688ms | 512.688ms | 512.688ms | 512.688ms |
| Hiveage | 2 | 493.102ms | 246.551ms | 223.123ms | 269.979ms |
| HubSpotApiKey | 2 | 491.264ms | 245.632ms | 110.336ms | 380.929ms |
| ProxyCrawl | 1 | 489.6ms | 489.6ms | 489.6ms | 489.6ms |
| ApiDeck | 1 | 486.053ms | 486.053ms | 486.053ms | 486.053ms |
| CustomerIO | 1 | 482ms | 482ms | 482ms | 482ms |
| Dyspatch | 1 | 466.547ms | 466.547ms | 466.547ms | 466.547ms |
| CalorieNinja | 1 | 465.953ms | 465.953ms | 465.953ms | 465.953ms |
| ClockworkSMS | 1 | 464.8ms | 464.8ms | 464.8ms | 464.8ms |
| Beebole | 1 | 460.823ms | 460.823ms | 460.823ms | 460.823ms |
| Coveralls | 1 | 459.602ms | 459.602ms | 459.602ms | 459.602ms |
| Happyscribe | 1 | 459.138ms | 459.138ms | 459.138ms | 459.138ms |
| ExchangeRatesAPI | 1 | 457.958ms | 457.958ms | 457.958ms | 457.958ms |
| Finnhub | 1 | 453.395ms | 453.395ms | 453.395ms | 453.395ms |
| FlightApi | 1 | 452.712ms | 452.712ms | 452.712ms | 452.712ms |
| Close | 1 | 448.435ms | 448.435ms | 448.435ms | 448.435ms |
| GeoIpifi | 1 | 445.2ms | 445.2ms | 445.2ms | 445.2ms |
| Browshot | 1 | 442.216ms | 442.216ms | 442.216ms | 442.216ms |
| ChecklyHQ | 1 | 439.258ms | 439.258ms | 439.258ms | 439.258ms |
| HolidayAPI | 1 | 434.202ms | 434.202ms | 434.202ms | 434.202ms |
| Currencyfreaks | 1 | 431.112ms | 431.112ms | 431.112ms | 431.112ms |
| ChecIO | 1 | 430.38ms | 430.38ms | 430.38ms | 430.38ms |
| CentralStationCRM | 1 | 426.277ms | 426.277ms | 426.277ms | 426.277ms |
| Diffbot | 1 | 425.263ms | 425.263ms | 425.263ms | 425.263ms |
| RechargePayments | 1 | 422.302ms | 422.302ms | 422.302ms | 422.302ms |
| Apacta | 1 | 417.636ms | 417.636ms | 417.636ms | 417.636ms |
| Axonaut | 1 | 416.885ms | 416.885ms | 416.885ms | 416.885ms |
| Geoapify | 1 | 416.807ms | 416.807ms | 416.807ms | 416.807ms |
| Budibase | 1 | 416.174ms | 416.174ms | 416.174ms | 416.174ms |
| Cashboard | 1 | 415.471ms | 415.471ms | 415.471ms | 415.471ms |
| AzureDevopsPersonalAccessToken | 1 | 406.777ms | 406.777ms | 406.777ms | 406.777ms |
| CloudflareCaKey | 2 | 406.184ms | 203.092ms | 187.931ms | 218.252ms |
| D7Network | 1 | 398.129ms | 398.129ms | 398.129ms | 398.129ms |
| Apilayer | 1 | 397.513ms | 397.513ms | 397.513ms | 397.513ms |
| Dovico | 1 | 390.418ms | 390.418ms | 390.418ms | 390.418ms |
| Graphhopper | 1 | 389.987ms | 389.987ms | 389.987ms | 389.987ms |
| FacePlusPlus | 1 | 387.464ms | 387.464ms | 387.464ms | 387.464ms |
| GetGeoAPI | 1 | 384.357ms | 384.357ms | 384.357ms | 384.357ms |
| CommerceJS | 1 | 378.765ms | 378.765ms | 378.765ms | 378.765ms |
| Autoklose | 1 | 375.879ms | 375.879ms | 375.879ms | 375.879ms |
| Documo | 1 | 375.695ms | 375.695ms | 375.695ms | 375.695ms |
| Github | 9 | 372.518ms | 41.391ms | 28.091ms | 81.574ms |
| Hypertrack | 1 | 371.135ms | 371.135ms | 371.135ms | 371.135ms |
| PaypalOauth | 1 | 358.992ms | 358.992ms | 358.992ms | 358.992ms |
| CryptoCompare | 1 | 358.242ms | 358.242ms | 358.242ms | 358.242ms |
| CloudImage | 1 | 354.744ms | 354.744ms | 354.744ms | 354.744ms |
| Borgbase | 1 | 353.087ms | 353.087ms | 353.087ms | 353.087ms |
| Getgist | 1 | 352.302ms | 352.302ms | 352.302ms | 352.302ms |
| ZipBooks | 1 | 351.694ms | 351.694ms | 351.694ms | 351.694ms |
| Geocodify | 1 | 348.178ms | 348.178ms | 348.178ms | 348.178ms |
| ElasticEmail | 1 | 338.959ms | 338.959ms | 338.959ms | 338.959ms |
| FXMarket | 1 | 336.454ms | 336.454ms | 336.454ms | 336.454ms |
| Fullstory | 3 | 328.701ms | 109.567ms | 75.551ms | 168.733ms |
| Coinlib | 1 | 328.328ms | 328.328ms | 328.328ms | 328.328ms |
| Deepgram | 1 | 327.849ms | 327.849ms | 327.849ms | 327.849ms |
| Ethplorer | 1 | 324.984ms | 324.984ms | 324.984ms | 324.984ms |
| Aylien | 1 | 324.339ms | 324.339ms | 324.339ms | 324.339ms |
| BoostNote | 1 | 323.283ms | 323.283ms | 323.283ms | 323.283ms |
| Glassnode | 1 | 321.283ms | 321.283ms | 321.283ms | 321.283ms |
| Doppler | 1 | 321.24ms | 321.24ms | 321.24ms | 321.24ms |
| DataGov | 1 | 317.356ms | 317.356ms | 317.356ms | 317.356ms |
| BrowserStack | 1 | 314.535ms | 314.535ms | 314.535ms | 314.535ms |
| EdenAI | 1 | 312.864ms | 312.864ms | 312.864ms | 312.864ms |
| CustomerGuru | 1 | 311.783ms | 311.783ms | 311.783ms | 311.783ms |
| Fetchrss | 1 | 309.581ms | 309.581ms | 309.581ms | 309.581ms |
| FormIO | 1 | 309.1ms | 309.1ms | 309.1ms | 309.1ms |
| GoCardless | 1 | 307.221ms | 307.221ms | 307.221ms | 307.221ms |
| Anypoint | 1 | 303.397ms | 303.397ms | 303.397ms | 303.397ms |
| Squareup | 1 | 302.707ms | 302.707ms | 302.707ms | 302.707ms |
| Dareboost | 1 | 301.817ms | 301.817ms | 301.817ms | 301.817ms |
| Cloudplan | 1 | 301.475ms | 301.475ms | 301.475ms | 301.475ms |
| Caflou | 1 | 299.866ms | 299.866ms | 299.866ms | 299.866ms |
| Chatbot | 1 | 297.884ms | 297.884ms | 297.884ms | 297.884ms |
| Api2Cart | 1 | 293.651ms | 293.651ms | 293.651ms | 293.651ms |
| GTMetrix | 1 | 293.491ms | 293.491ms | 293.491ms | 293.491ms |
| Clockify | 1 | 289.411ms | 289.411ms | 289.411ms | 289.411ms |
| DeepAI | 1 | 287.12ms | 287.12ms | 287.12ms | 287.12ms |
| CloudElements | 1 | 286.963ms | 286.963ms | 286.963ms | 286.963ms |
| Everhour | 1 | 286.541ms | 286.541ms | 286.541ms | 286.541ms |
| Anthropic | 4 | 284.898ms | 71.224ms | 48.746ms | 93.977ms |
| Front | 1 | 284.225ms | 284.225ms | 284.225ms | 284.225ms |
| Guardianapi | 1 | 283.843ms | 283.843ms | 283.843ms | 283.843ms |
| Tailscale | 1 | 280.093ms | 280.093ms | 280.093ms | 280.093ms |
| Ditto | 1 | 276.451ms | 276.451ms | 276.451ms | 276.451ms |
| Nightfall | 1 | 273.092ms | 273.092ms | 273.092ms | 273.092ms |
| Appcues | 1 | 267.501ms | 267.501ms | 267.501ms | 267.501ms |
| Column | 1 | 267.317ms | 267.317ms | 267.317ms | 267.317ms |
| AsanaPersonalAccessToken | 1 | 267.045ms | 267.045ms | 267.045ms | 267.045ms |
| BlitApp | 1 | 263.727ms | 263.727ms | 263.727ms | 263.727ms |
| FigmaPersonalAccessToken | 2 | 260.488ms | 130.244ms | 128.723ms | 131.765ms |
| Hybiscus | 1 | 258.084ms | 258.084ms | 258.084ms | 258.084ms |
| Dockerhub | 4 | 246.728ms | 61.682ms | 19.292ms | 102.75ms |
| CoinbaseWaaS | 1 | 244.883ms | 244.883ms | 244.883ms | 244.883ms |
| Freshbooks | 1 | 242.631ms | 242.631ms | 242.631ms | 242.631ms |
| Checkout | 1 | 238.762ms | 238.762ms | 238.762ms | 238.762ms |
| Verifier | 1 | 235.604ms | 235.604ms | 235.604ms | 235.604ms |
| Detectify | 1 | 234.516ms | 234.516ms | 234.516ms | 234.516ms |
| ClickHelp | 1 | 233.635ms | 233.635ms | 233.635ms | 233.635ms |
| Brandfetch | 1 | 232.248ms | 232.248ms | 232.248ms | 232.248ms |
| Chartmogul | 1 | 231.784ms | 231.784ms | 231.784ms | 231.784ms |
| EnvoyApiKey | 1 | 231.549ms | 231.549ms | 231.549ms | 231.549ms |
| Mailchimp | 1 | 231.017ms | 231.017ms | 231.017ms | 231.017ms |
| SourcegraphCody | 1 | 229.985ms | 229.985ms | 229.985ms | 229.985ms |
| Bulbul | 1 | 227.823ms | 227.823ms | 227.823ms | 227.823ms |
| AssemblyAI | 1 | 223.569ms | 223.569ms | 223.569ms | 223.569ms |
| Stripe | 3 | 221.686ms | 73.895ms | 8.246ms | 192.671ms |
| Confluent | 1 | 218.351ms | 218.351ms | 218.351ms | 218.351ms |
| Dandelion | 1 | 216.691ms | 216.691ms | 216.691ms | 216.691ms |
| Duply | 1 | 210.589ms | 210.589ms | 210.589ms | 210.589ms |
| AzureStorage | 3 | 209.984ms | 69.995ms | 12.606ms | 173.825ms |
| Grafana | 2 | 205.315ms | 102.657ms | 90.281ms | 115.034ms |
| Shopify | 1 | 199.481ms | 199.481ms | 199.481ms | 199.481ms |
| ClickupPersonalToken | 1 | 197.504ms | 197.504ms | 197.504ms | 197.504ms |
| CalendlyApiKey | 1 | 196.445ms | 196.445ms | 196.445ms | 196.445ms |
| FlowFlu | 1 | 196.325ms | 196.325ms | 196.325ms | 196.325ms |
| FinancialModelingPrep | 1 | 195.293ms | 195.293ms | 195.293ms | 195.293ms |
| Feedier | 1 | 192.816ms | 192.816ms | 192.816ms | 192.816ms |
| EagleEyeNetworks | 1 | 188.287ms | 188.287ms | 188.287ms | 188.287ms |
| GrafanaServiceAccount | 1 | 187.365ms | 187.365ms | 187.365ms | 187.365ms |
| Paystack | 1 | 174.902ms | 174.902ms | 174.902ms | 174.902ms |
| Dfuse | 1 | 166.846ms | 166.846ms | 166.846ms | 166.846ms |
| NpmToken | 1 | 165.597ms | 165.597ms | 165.597ms | 165.597ms |
| Etherscan | 1 | 157.354ms | 157.354ms | 157.354ms | 157.354ms |
| ExportSDK | 1 | 156.952ms | 156.952ms | 156.952ms | 156.952ms |
| Codemagic | 1 | 154.157ms | 154.157ms | 154.157ms | 154.157ms |
| Censys | 1 | 152.572ms | 152.572ms | 152.572ms | 152.572ms |
| Docusign | 1 | 149.578ms | 149.578ms | 149.578ms | 149.578ms |
| BetterStack | 1 | 147.852ms | 147.852ms | 147.852ms | 147.852ms |
| Float | 1 | 147.837ms | 147.837ms | 147.837ms | 147.837ms |
| Codacy | 1 | 147.024ms | 147.024ms | 147.024ms | 147.024ms |
| BscScan | 1 | 145.279ms | 145.279ms | 145.279ms | 145.279ms |
| Freshdesk | 1 | 141.709ms | 141.709ms | 141.709ms | 141.709ms |
| Apptivo | 1 | 140.472ms | 140.472ms | 140.472ms | 140.472ms |
| Ramp | 1 | 139.639ms | 139.639ms | 139.639ms | 139.639ms |
| ReadMe | 1 | 134.913ms | 134.913ms | 134.913ms | 134.913ms |
| Docparser | 1 | 127.008ms | 127.008ms | 127.008ms | 127.008ms |
| Databox | 1 | 125.364ms | 125.364ms | 125.364ms | 125.364ms |
| Dnscheck | 1 | 123.67ms | 123.67ms | 123.67ms | 123.67ms |
| Parsehub | 1 | 122.952ms | 122.952ms | 122.952ms | 122.952ms |
| SendGrid | 1 | 122.495ms | 122.495ms | 122.495ms | 122.495ms |
| Cloverly | 1 | 122.485ms | 122.485ms | 122.485ms | 122.485ms |
| AWSSessionKey | 2 | 121.902ms | 60.951ms | 26.838ms | 95.063ms |
| ZulipChat | 1 | 121.735ms | 121.735ms | 121.735ms | 121.735ms |
| Coda | 1 | 121.071ms | 121.071ms | 121.071ms | 121.071ms |
| Geocode | 1 | 120.728ms | 120.728ms | 120.728ms | 120.728ms |
| BraintreePayments | 1 | 120.507ms | 120.507ms | 120.507ms | 120.507ms |
| Dropbox | 1 | 120.469ms | 120.469ms | 120.469ms | 120.469ms |
| AutoPilot | 1 | 120.201ms | 120.201ms | 120.201ms | 120.201ms |
| LinearAPI | 1 | 118.711ms | 118.711ms | 118.711ms | 118.711ms |
| Notion | 1 | 118.32ms | 118.32ms | 118.32ms | 118.32ms |
| Atera | 1 | 116.39ms | 116.39ms | 116.39ms | 116.39ms |
| GoodDay | 1 | 115.271ms | 115.271ms | 115.271ms | 115.271ms |
| Getresponse | 1 | 114.889ms | 114.889ms | 114.889ms | 114.889ms |
| CloudConvert | 1 | 114.536ms | 114.536ms | 114.536ms | 114.536ms |
| ConversionTools | 1 | 113.072ms | 113.072ms | 113.072ms | 113.072ms |
| DiscordWebhook | 1 | 112.658ms | 112.658ms | 112.658ms | 112.658ms |
| FastForex | 1 | 111.693ms | 111.693ms | 111.693ms | 111.693ms |
| Onedesk | 1 | 107.935ms | 107.935ms | 107.935ms | 107.935ms |
| RubyGems | 1 | 107.805ms | 107.805ms | 107.805ms | 107.805ms |
| Cloudmersive | 1 | 107.282ms | 107.282ms | 107.282ms | 107.282ms |
| ExtractorAPI | 1 | 106.193ms | 106.193ms | 106.193ms | 106.193ms |
| LaunchDarkly | 2 | 106.07ms | 53.035ms | 27.001ms | 79.069ms |
| Eventbrite | 1 | 104.769ms | 104.769ms | 104.769ms | 104.769ms |
| Kontent | 2 | 101.687ms | 50.844ms | 50.378ms | 51.309ms |
| Demio | 1 | 100.384ms | 100.384ms | 100.384ms | 100.384ms |
| Chatfule | 1 | 99.389ms | 99.389ms | 99.389ms | 99.389ms |
| DiscordBotToken | 1 | 97.597ms | 97.597ms | 97.597ms | 97.597ms |
| FourSquare | 1 | 97.366ms | 97.366ms | 97.366ms | 97.366ms |
| Clarifai | 1 | 96.088ms | 96.088ms | 96.088ms | 96.088ms |
| FlatIO | 1 | 94.507ms | 94.507ms | 94.507ms | 94.507ms |
| SatismeterWritekey | 1 | 94.283ms | 94.283ms | 94.283ms | 94.283ms |
| Besnappy | 1 | 93.116ms | 93.116ms | 93.116ms | 93.116ms |
| EcoStruxureIT | 1 | 93.069ms | 93.069ms | 93.069ms | 93.069ms |
| Buildkite | 2 | 91.863ms | 45.931ms | 31.326ms | 60.536ms |
| ExchangeRateAPI | 2 | 91.397ms | 45.698ms | 41.348ms | 50.049ms |
| Clinchpad | 1 | 88.634ms | 88.634ms | 88.634ms | 88.634ms |
| Blazemeter | 1 | 88.316ms | 88.316ms | 88.316ms | 88.316ms |
| Geckoboard | 1 | 87.878ms | 87.878ms | 87.878ms | 87.878ms |
| Artsy | 1 | 86.878ms | 86.878ms | 86.878ms | 86.878ms |
| Enigma | 1 | 85.711ms | 85.711ms | 85.711ms | 85.711ms |
| Host | 1 | 80.31ms | 80.31ms | 80.31ms | 80.31ms |
| Cicero | 1 | 78.429ms | 78.429ms | 78.429ms | 78.429ms |
| Sourcegraph | 1 | 78.365ms | 78.365ms | 78.365ms | 78.365ms |
| Groovehq | 1 | 78.342ms | 78.342ms | 78.342ms | 78.342ms |
| Pulumi | 1 | 78.17ms | 78.17ms | 78.17ms | 78.17ms |
| Caspio | 1 | 77.56ms | 77.56ms | 77.56ms | 77.56ms |
| Audd | 1 | 77.128ms | 77.128ms | 77.128ms | 77.128ms |
| Bugsnag | 1 | 77.076ms | 77.076ms | 77.076ms | 77.076ms |
| DigitalOceanToken | 1 | 76.647ms | 76.647ms | 76.647ms | 76.647ms |
| FixerIO | 1 | 75.504ms | 75.504ms | 75.504ms | 75.504ms |
| Flickr | 1 | 75.217ms | 75.217ms | 75.217ms | 75.217ms |
| MicrosoftTeamsWebhook | 1 | 74.966ms | 74.966ms | 74.966ms | 74.966ms |
| Fulcrum | 1 | 74.75ms | 74.75ms | 74.75ms | 74.75ms |
| Baremetrics | 1 | 73.959ms | 73.959ms | 73.959ms | 73.959ms |
| Cloudsmith | 1 | 73.92ms | 73.92ms | 73.92ms | 73.92ms |
| Beamer | 1 | 73.221ms | 73.221ms | 73.221ms | 73.221ms |
| Gemini | 1 | 72.986ms | 72.986ms | 72.986ms | 72.986ms |
| FrameIO | 1 | 71.897ms | 71.897ms | 71.897ms | 71.897ms |
| AirtableApiKey | 1 | 70.165ms | 70.165ms | 70.165ms | 70.165ms |
| Gyazo | 1 | 67.666ms | 67.666ms | 67.666ms | 67.666ms |
| Gumroad | 1 | 66.396ms | 66.396ms | 66.396ms | 66.396ms |
| BombBomb | 1 | 65.728ms | 65.728ms | 65.728ms | 65.728ms |
| Cliengo | 1 | 65.027ms | 65.027ms | 65.027ms | 65.027ms |
| Ubidots | 1 | 65.027ms | 65.027ms | 65.027ms | 65.027ms |
| ButterCMS | 1 | 64.865ms | 64.865ms | 64.865ms | 64.865ms |
| Convertkit | 1 | 64.721ms | 64.721ms | 64.721ms | 64.721ms |
| Circle | 1 | 64.202ms | 64.202ms | 64.202ms | 64.202ms |
| DatadogToken | 1 | 63.882ms | 63.882ms | 63.882ms | 63.882ms |
| Harvest | 1 | 63.805ms | 63.805ms | 63.805ms | 63.805ms |
| BitcoinAverage | 1 | 63.54ms | 63.54ms | 63.54ms | 63.54ms |
| Coinbase | 1 | 63.149ms | 63.149ms | 63.149ms | 63.149ms |
| Delighted | 1 | 62.955ms | 62.955ms | 62.955ms | 62.955ms |
| Apiflash | 1 | 62.863ms | 62.863ms | 62.863ms | 62.863ms |
| ConvertApi | 1 | 62.473ms | 62.473ms | 62.473ms | 62.473ms |
| Okta | 1 | 62.098ms | 62.098ms | 62.098ms | 62.098ms |
| Courier | 1 | 62.084ms | 62.084ms | 62.084ms | 62.084ms |
| Bitbar | 1 | 61.583ms | 61.583ms | 61.583ms | 61.583ms |
| DatabricksToken | 1 | 60.955ms | 60.955ms | 60.955ms | 60.955ms |
| Ipapi | 1 | 60.541ms | 60.541ms | 60.541ms | 60.541ms |
| Blogger | 1 | 59.057ms | 59.057ms | 59.057ms | 59.057ms |
| FacebookOAuth | 1 | 58.393ms | 58.393ms | 58.393ms | 58.393ms |
| GoogleOauth2 | 1 | 56.618ms | 56.618ms | 56.618ms | 56.618ms |
| Postman | 1 | 56.403ms | 56.403ms | 56.403ms | 56.403ms |
| Gengo | 1 | 56.281ms | 56.281ms | 56.281ms | 56.281ms |
| Heroku | 1 | 56.19ms | 56.19ms | 56.19ms | 56.19ms |
| Slack | 1 | 56.171ms | 56.171ms | 56.171ms | 56.171ms |
| Findl | 1 | 56.104ms | 56.104ms | 56.104ms | 56.104ms |
| CoinApi | 1 | 55.934ms | 55.934ms | 55.934ms | 55.934ms |
| AsanaOauth | 1 | 54.194ms | 54.194ms | 54.194ms | 54.194ms |
| LDAP | 2 | 54.042ms | 27.021ms | 15.169ms | 38.874ms |
| Disqus | 1 | 53.743ms | 53.743ms | 53.743ms | 53.743ms |
| Bannerbear | 1 | 52.793ms | 52.793ms | 52.793ms | 52.793ms |
| DetectLanguage | 1 | 51.629ms | 51.629ms | 51.629ms | 51.629ms |
| LocationIQ | 1 | 50.433ms | 50.433ms | 50.433ms | 50.433ms |
| TwitterConsumerkey | 1 | 48.933ms | 48.933ms | 48.933ms | 48.933ms |
| Autodesk | 1 | 48.833ms | 48.833ms | 48.833ms | 48.833ms |
| SupabaseToken | 1 | 48.217ms | 48.217ms | 48.217ms | 48.217ms |
| Codeclimate | 1 | 46.116ms | 46.116ms | 46.116ms | 46.116ms |
| Humanity | 1 | 44.768ms | 44.768ms | 44.768ms | 44.768ms |
| Twilio | 1 | 43.892ms | 43.892ms | 43.892ms | 43.892ms |
| HuggingFace | 1 | 41.263ms | 41.263ms | 41.263ms | 41.263ms |
| Flightstats | 1 | 40.583ms | 40.583ms | 40.583ms | 40.583ms |
| Copper | 1 | 40.409ms | 40.409ms | 40.409ms | 40.409ms |
| Crowdin | 1 | 39.886ms | 39.886ms | 39.886ms | 39.886ms |
| Edamam | 1 | 39.239ms | 39.239ms | 39.239ms | 39.239ms |
| Prefect | 1 | 39.199ms | 39.199ms | 39.199ms | 39.199ms |
| Coinlayer | 1 | 38.744ms | 38.744ms | 38.744ms | 38.744ms |
| PlanetScaleDb | 1 | 38.076ms | 38.076ms | 38.076ms | 38.076ms |
| Apify | 1 | 38.06ms | 38.06ms | 38.06ms | 38.06ms |
| Hunter | 1 | 36.546ms | 36.546ms | 36.546ms | 36.546ms |
| CountryLayer | 1 | 36.021ms | 36.021ms | 36.021ms | 36.021ms |
| DroneCI | 1 | 35.969ms | 35.969ms | 35.969ms | 35.969ms |
| RabbitMQ | 1 | 35.807ms | 35.807ms | 35.807ms | 35.807ms |
| TerraformCloudPersonalToken | 1 | 35.705ms | 35.705ms | 35.705ms | 35.705ms |
| MapBox | 1 | 32.878ms | 32.878ms | 32.878ms | 32.878ms |
| Convier | 1 | 31.93ms | 31.93ms | 31.93ms | 31.93ms |
| Uclassify | 1 | 31.691ms | 31.691ms | 31.691ms | 31.691ms |
| CapsuleCRM | 1 | 31.584ms | 31.584ms | 31.584ms | 31.584ms |
| HereAPI | 1 | 30.35ms | 30.35ms | 30.35ms | 30.35ms |
| DronaHQ | 1 | 29.8ms | 29.8ms | 29.8ms | 29.8ms |
| AvazaPersonalAccessToken | 1 | 29.241ms | 29.241ms | 29.241ms | 29.241ms |
| BitLyAccessToken | 1 | 29.228ms | 29.228ms | 29.228ms | 29.228ms |
| HelloSign | 1 | 28.341ms | 28.341ms | 28.341ms | 28.341ms |
| PlanetScale | 1 | 27.847ms | 27.847ms | 27.847ms | 27.847ms |
| Salesforce | 1 | 27.426ms | 27.426ms | 27.426ms | 27.426ms |
| Gitter | 1 | 26.08ms | 26.08ms | 26.08ms | 26.08ms |
| Helpscout | 1 | 24.519ms | 24.519ms | 24.519ms | 24.519ms |
| Honeycomb | 1 | 21.603ms | 21.603ms | 21.603ms | 21.603ms |
| BlockNative | 1 | 21.327ms | 21.327ms | 21.327ms | 21.327ms |
| Formsite | 1 | 19.28ms | 19.28ms | 19.28ms | 19.28ms |
| AirbrakeUserKey | 1 | 19.089ms | 19.089ms | 19.089ms | 19.089ms |
| NGC | 1 | 18.69ms | 18.69ms | 18.69ms | 18.69ms |
| Bugherd | 1 | 15.435ms | 15.435ms | 15.435ms | 15.435ms |
| SlackWebhook | 1 | 13.847ms | 13.847ms | 13.847ms | 13.847ms |
| AzureSearchQueryKey | 1 | 10.837ms | 10.837ms | 10.837ms | 10.837ms |
| TinesWebhook | 1 | 10.829ms | 10.829ms | 10.829ms | 10.829ms |
| SQLServer | 3 | 9.604ms | 3.201ms | 452µs | 7.222ms |
| GCP | 2 | 623µs | 312µs | 186µs | 438µs |
| TrufflehogEnterprise | 1 | 188µs | 188µs | 188µs | 188µs |

## Analysis

The slowest detector is **JDBC** with a total execution time of 2m30.138951s across 18 calls.
The top 3 detectors account for 53.3% of total detection time.
