# relocale

## Installation

Before you start, you need to make sure that the `relocale` binary is installed
somewhere. I will leave that part up to you.

### Preparation

Create the `relocale` user:

```
useradd --system --shell /bin/false relocale
```

### Locale configuration

An [example configuration file](relocale.toml) is provided in this repository.

### Service configuration

Next, put the [systemd unit file](relocale@.service) into `/etc/systemd/system`.
This unit file might not work for every setup out there, so make sure to modify
it as needed.

### Reverse proxy

Now it's time to configure the web server. There is an
[example nginx configuration](nginx.conf) to get started.
