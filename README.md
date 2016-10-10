```
git daemon --export-all --base-path=/var/mirrord --max-connections=0 --reuseaddr
sudo iptables -t nat -A OUTPUT -p tcp --dport 80 -j DNAT --to-destination 127.0.0.1:80
```
