About
=======

Probe Linux, Windows, and ESXi for CPU and memory utilization, and push the results to a Pushgateway.

Background
-----------

We have a large num. of servers running in our lab, but engineers always compain there are not enough available resources although every one reserves many servers. This application probes servers for CPU and memory utilization based on OS stats and reports the results to a Prometheus Pushgateway for continuous monitoring (after integrating Prometheus and Grafana) - if the CPU and memory utilization on some servers are always low, the servers should be released for people who really need resources.

Prerequisites
--------------

- Allow Pings;
- Allow connections to ports (or just stop firewall):

  * Linux: 22;
  * Windows: 3389 (for OS dection), 5985 (for WinRM HTTP);
  * ESXi: 902 (for OS dection), 443 (for vSphere API).

- Linux: Password based ssh access;
- Windows:

  * A valid local credentail (domain credentials do not work);
  * Enable WinRM with basic auth:

    ::

      winrm quickconfig
      y
      winrm set winrm/config/service/Auth '@{Basic="true"}'
      winrm set winrm/config/service '@{AllowUnencrypted="true"}'
      winrm set winrm/config/winrs '@{MaxMemoryPerShellMB="1024"}'

- ESXi: Configure a valid password for access.

Usage
------

::

  cd scanner
  go build .
  cp hosts.json hosts.real.json
  vim hosts.test.json # Add your server IPs/FQDNs
  cp credentials.json credentials.test.json
  vim credentials.test.json # Define server access credentials
  ./scanner -h
  ./scanner -s hosts.test.json -p credentials.test.json -o servers.test.json
  cd ..
  go build .
  ./osprobe -h
  ./osprobe -c scanner/servers.test.json -g http://<pushgateway>:<port> -i <update interval>
