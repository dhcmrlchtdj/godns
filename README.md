# GoDNS

DNS server

## Example

```
$ go build
$ ./godns --help
$ ./godns --port=1053 --conf=/path/to/config
```

## Usage

### arch
```
$ git clone
$ cd aur && makepkg -srci
$ cat /etc/godns/config.json
$ systemctl enable --now godns.service
```

### mac
```
$ brew tap dhcmrlchtdj/godns https://github.com/dhcmrlchtdj/godns
$ brew install --HEAD dhcmrlchtdj/godns/godns
$ cat "$(brew --prefix)/etc/godns/config.json"
$ brew services start godns
```

## Config

```json
{
    "port": 1053,
    "log_level": "info",
    "rule": [
        {
            "pattern": { "domain": ["localhost"], "record": "A" },
            "upstream": { "ipv4": "127.0.0.1" }
        },
        {
            "pattern": { "suffix": ["."] },
            "upstream": {
                "doh": "https://1.1.1.1/dns-query",
                "doh_proxy": "http://127.0.0.1:1080"
            }
        },
        {
            "pattern": { "suffix": [".cn"] },
            "upstream": { "udp": "119.29.29.29:53" }
        }
    ]
}
```
