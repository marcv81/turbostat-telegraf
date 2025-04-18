# turbostat-telegraf

`turbostat-telegraf` is a Telegraf plugin to monitor system performance using `turbostat`.

## Installation

Build and install the `turbostat-telegraf` executable.

    go build ./
    sudo cp turbostat-telegraf /usr/local/bin/

Allow the `telegraf` user to run `turbostat-telegraf` in `/etc/sudoers.d/telegraf`.

    telegraf ALL=(root) NOPASSWD: /usr/local/bin/turbostat-telegraf

Create `/etc/telegraf/telegraf.d/turbostat.conf` with the following contents.

    [[inputs.execd]]
      command = ["sudo", "turbostat-telegraf", "--quiet", "-i", "10"]
      signal = "none"
      data_format = "influx"

Restart `telegraf`.
