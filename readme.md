### Erase Storage

```
sudo sed -i 's/databases 16/databases 81/g' /etc/redis/redis.conf
```

### Erase Disk Redis key
* starttasktime   
* endtasktime    
* errorcode 
* processing   

```

                     This      All      All     This                              Single  
Pass No. of          Pass   Passes   Passes     Pass              Est.     MB/      MB/   
 No. Passes Byte Complete Complete  Elapsed  Consume    Start   Finish   Second   Second  
---- ------ ---- -------- -------- -------- -------- -------- -------- -------- --------  
   1      1 0xff   0.159%   0.159% 00:00:05 00:00:05 10:38:01 00003141    93.18    93.18  

```
     paser this data to DataBase,  regex=(\d*\.\d*)%.*?(\d*\.\d*)%.*?(\d*\.\d*)$

|name      |       index|
|:-------|---------:|
|speed|9|
|start|7|
|time|5|
|est|8|
|progress|4|
|optime|time/est|


# Error code List
### base errorcode 36000, all error code +36000
|error code|means|
|----------|-----|
|999|user cancel task|
|100|sanitize not support|
|123|unknow failed, not start erasing|
|250|user input disk transcation|
|10|sanitize verify failed|
|11|not find linuxname|
|12|not find sgName|
|13|create log file failed|


```
this Model write buffer size is small
using buffer size 200 is tested

sudo smartctl -a /dev/sg13
[sudo] password for dsed: 
smartctl 7.1 2019-12-30 r5022 [x86_64-linux-5.11.0-34-generic] (local build)
Copyright (C) 2002-19, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Vendor:               SEAGATE
Product:              ST1800MM0008
Revision:             K004
Compliance:           SPC-4
User Capacity:        1,800,360,124,416 bytes [1.80 TB]
Logical block size:   4096 bytes
LU is fully provisioned
Rotation Rate:        10500 rpm
Form Factor:          2.5 inches
Logical Unit id:      0x5000c5009eae245f
Serial number:        W3Z03T6V0000E729K75N
Device type:          disk
Transport protocol:   SAS (SPL-3)
Local Time is:        Wed Sep 29 16:17:59 2021 PDT
SMART support is:     Available - device has SMART capability.
SMART support is:     Enabled
Temperature Warning:  Enabled

=== START OF READ SMART DATA SECTION ===
SMART Health Status: OK

Grown defects during certification <not available>
Total blocks reassigned during format <not available>
Total new blocks reassigned <not available>
Power on minutes since format <not available>
Current Drive Temperature:     56 C
Drive Trip Temperature:        60 C

Manufactured in week 06 of year 2017
Specified cycle count over device lifetime:  10000
Accumulated start-stop cycles:  137
Specified load-unload count over device lifetime:  300000
Accumulated load-unload cycles:  1428
Elements in grown defect list: 0

Vendor (Seagate Cache) information
  Blocks sent to initiator = 12911371
  Blocks received from initiator = 1328554629
  Blocks read from cache and sent to initiator = 19566
  Number of read and write commands whose size <= segment size = 33229
  Number of read and write commands whose size > segment size = 0

Vendor (Seagate/Hitachi) factory information
  number of hours powered up = 31145.02
  number of minutes until next internal SMART test = 28

Error counter log:
           Errors Corrected by           Total   Correction     Gigabytes    Total
               ECC          rereads/    errors   algorithm      processed    uncorrected
           fast | delayed   rewrites  corrected  invocations   [10^9 bytes]  errors
read:   25822856        0         0  25822856          0         52.885           0
write:         0        0         0         0          0       5441.840           0

Non-medium error count:        0


[GLTSD (Global Logging Target Save Disable) set. Enable Save with '-S on']
SMART Self-test log
Num  Test              Status                 segment  LifeTime  LBA_first_err [SK ASC ASQ]
     Description                              number   (hours)
# 1  Background short  Completed                   -      11                 - [-   -    -]
# 2  Background short  Completed                   -      10                 - [-   -    -]
# 3  Background short  Completed                   -       8                 - [-   -    -]
# 4  Background short  Completed                   -       5                 - [-   -    -]
# 5  Background short  Completed                   -       3                 - [-   -    -]

Long (extended) Self-test duration: 11350 seconds [189.2 minutes]

```