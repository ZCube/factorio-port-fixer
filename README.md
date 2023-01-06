# factorio-port-fixer

## Matching Server Port fixer

## Usage

* Add /etc/hosts or c:\windows\system32\drivers\etc\hosts
  ```
  144.24.94.63 pingpong1.factorio.com
  144.24.94.63 pingpong2.factorio.com
  144.24.94.63 pingpong3.factorio.com
  144.24.94.63 pingpong4.factorio.com
  ```
* Add extra_hosts (docker-compose)
  ```
  extra_hosts:
    - 'pingpong1.factorio.com:144.24.94.63'
    - 'pingpong2.factorio.com:144.24.94.63'
    - 'pingpong3.factorio.com:144.24.94.63'
    - 'pingpong4.factorio.com:144.24.94.63'
  ```

* 144.24.94.63 is sample server.

## Example

* local mode (with ipify)

```yaml
version: '3.5'

services:
  pingpong:
    image: ghcr.io/zcube/factorio-port-fixer
    command: /factorio-port-fixer local --ip=0.0.0.0 --port=34197 --remotePort=${PORT:-34197}
    restart: unless-stopped
    environment:
      - PORT=${PORT:-34197}
    logging:
      driver: "json-file"
      options:
        max-size: "1m"

  factorio:
    image: factoriotools/factorio
    restart: unless-stopped
    environment:
      - PORT=${PORT:-34197}
    ports:
     - "${PORT:-34197}:${PORT:-34197}/udp"
     - "27015:27015/tcp"
    volumes:
     - /etc/localtime:/etc/localtime:ro
     - ./factorio:/factorio
    environment:
     - TZ=UTC
    links:
      - 'pingpong:pingpong1.factorio.com'
      - 'pingpong:pingpong2.factorio.com'
      - 'pingpong:pingpong3.factorio.com'
      - 'pingpong:pingpong4.factorio.com'
    # ping check
    healthcheck:
      test: curl --fail pingpong:34197/health || exit 1
      interval: 20s
      retries: 5
      start_period: 20s
      timeout: 10s
```

* remote mode

```yaml
version: '3.5'

services:
  factorio:
    image: factoriotools/factorio
    restart: unless-stopped
    ports:
     - "31497:31497/udp"
     - "27015:27015/tcp"
    volumes:
     - /etc/localtime:/etc/localtime:ro
     - ./factorio:/factorio
    environment:
     - TZ=UTC
    # sample server on oci remote port 34197 fixed
    extra_hosts:
     - 'pingpong1.factorio.com:144.24.94.63'
     - 'pingpong2.factorio.com:144.24.94.63'
     - 'pingpong3.factorio.com:144.24.94.63'
     - 'pingpong4.factorio.com:144.24.94.63'
    # ping check
    healthcheck:
      test: curl --fail pingpong1.factorio.com:34197/health || exit 1
      interval: 20s
      retries: 5
      start_period: 20s
      timeout: 10s
```

* pingpong server for remote mode (144.24.94.63)
```yaml
version: '3'
services:
  factorio:
    image: ghcr.io/zcube/factorio-port-fixer
    restart: unless-stopped
    ports:
     - "34197:34197/udp"
     - "34197:34197/tcp"
```
## License

MIT License

Copyright (c) 2023 ZCube

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
