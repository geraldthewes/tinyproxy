A very simple http/https proxy used to debug services

Summary
--------

Basically the ideas is:

Client <---> tinyproxy <---> Single remote site

Amazingly could not find such a utility.

The traffic betyween both is logged (for debugging purposes). This is meant for development only.

Instructions
------------

To use do

```
./tinyproxy -port 9999 -remote https://www.apple.com:443
```

Then:

```
curl -v http://127.0.0.1:9999/index.html
```
