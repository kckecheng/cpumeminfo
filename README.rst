About
=======

Probe Linux, Windows, and ESXi for CPU and memory utilization, and push the results to a Pushgateway.

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
