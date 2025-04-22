# turbostat-telegraf-plugin

This [Telegraf](https://github.com/influxdata/telegraf) plugin monitors system performance using `turbostat`.

## Installation

Install `turbostat`.

Build and install the `turbostat-telegraf-plugin` executable.

    go build ./cmd/turbostat-telegraf-plugin/
    sudo cp turbostat-telegraf-plugin /usr/local/bin/

Allow the `telegraf` user to run `turbostat` in `/etc/sudoers.d/telegraf`.

    telegraf ALL=(root) NOPASSWD: /usr/bin/turbostat

Create `/etc/telegraf/telegraf.d/turbostat.conf` with the following contents.

    [[inputs.execd]]
      command = ["turbostat-telegraf-plugin", "sudo", "turbostat", "--quiet", "-i", "10"]
      signal = "none"
      data_format = "influx"


Restart `telegraf`.
