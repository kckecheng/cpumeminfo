About
=======

Probe Linux, Windows, and ESXi for CPU and memory utilization, and push the results to a Pushgateway.

Background
-----------

We have a large num. of servers running in our lab, but engineers always compain there are not enough available resources although every one reserves many servers. This application probes servers for CPU and memory utilization based on OS stats and reports the results to a Prometheus Pushgateway for continuous monitoring (after integrating Prometheus and Grafana) - if the CPU and memory utilization on some servers are always low, the servers should be released for people who really need resources.

Prerequisites
--------------

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

  go build .
  ./osprobe -h
