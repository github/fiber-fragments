# Fragments Middleware

```bash
echo "GET http://localhost:8080/index" | vegeta attack -duration=5s -rate 1000 | tee results.bin | vegeta report
  vegeta report -type=json results.bin > metrics.json
  cat results.bin | vegeta plot > plot.html
  cat results.bin | vegeta report -type="hist[0,100ms,200ms,300ms]"

Requests      [total, rate, throughput]         5000, 1000.22, 995.34
Duration      [total, attack, wait]             5.023s, 4.999s, 24.486ms
Latencies     [min, mean, 50, 90, 95, 99, max]  24.007ms, 28.534ms, 24.816ms, 28.487ms, 31.818ms, 139.57ms, 155.654ms
Bytes In      [total, mean]                     2755000, 551.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:5000
Error Set:
Bucket           #     %       Histogram
[0s,     100ms]  4866  97.32%  ########################################################################
[100ms,  200ms]  134   2.68%   ##
[200ms,  300ms]  0     0.00%
[300ms,  +Inf]   0     0.00%
```