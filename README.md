# relocale

## Overview

relocale redirects requests via the Accept-Language HTTP header. It is meant
for static sites that need to redirect users based on their locale without any
JavaScript or meta tags.

## Installation

Before you start, you need to make sure that the `relocale` binary is installed
somewhere. I will leave that part up to you.

### Preparation

Create the `relocale` user:

```
useradd --system --shell /bin/false relocale
```

### Locale configuration

Here is an example of how a configuration file could be written:

```toml
# Mandatory fallbacks which are used when no or empty values have been provided.
locale = 'en'
destination = '/${locale}${path}'

[network]
# Network address for relocale.
# This defaults to localhost:10451.
address = 'localhost:10451'

[locales]
# This regular expression will set the locale to 'en' if the client's locale
# matches 'en', 'en-US', ...
en.alias = '^en-.+$'

de.alias = '^de-.+$'

# The destination path which will be redirected to.
# This is a template string with the following predefined parameters:
# * path: Original request path
# * locale: Current locale
de.destination = '/${locale}${path}'
```

### Service configuration

```ini
[Unit]
Description=relocale redirection service
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/relocale --config %i
ProtectSystem=strict
ReadOnlyPaths=/etc/relocale
WorkingDirectory=/etc/relocale
TimeoutStopSec=20
KillMode=mixed
Restart=on-failure
User=relocale
Group=relocale

[Install]
WantedBy=multi-user.target
```

Next, put the systemd unit file (e.g. as `relocale@.service`) into
`/etc/systemd/system`. This unit file might not work for every setup out there,
so make sure to modify it as needed.

### Reverse proxy

```nginx
server {
    # Change this to match your server's name or domain.
    server_name example.com;

    # This matches any path without the available locales 'en' and 'de' at the
    # start.
    location ~ ^/(?!(de|en)) {
        # Change this to match relocale's configuration.
        proxy_pass http://localhost:10451;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Every other request will serve a static file.
    location / {
        root /srv/http/example.com;
        index index.html;
    }

    # Your TLS configuration goes here.
    listen 80;
}
```
