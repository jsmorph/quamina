# A performance test tool

_Work in progress_

This tool generates events and patterns, and then the tool can measure
performance of various Quamina operations.

A few modes (`-mode`):

1. `events`: Time `MatchesForEvent`
1. `patterns`: Time `AddPattern`
1. `showgen`: Just print the generated events and patterns
1. `concurrent`: A concurrency test of `MatchesForEvent`


## Example: `patterns`

```Shell
quaperf -mode patterns -patterns 5000
```

```
2022/11/06 23:06:54 generated 5000 patterns and 10000 events in 1.331549843s
patterns,msPerPattern,msPerPatternTail
200,0.144779,0.144777
400,0.315426,0.481844
600,0.410396,0.594660
800,0.531183,0.893438
1000,0.625418,1.002236
1200,0.723858,1.215962
1400,0.850897,1.613043
1600,0.953209,1.669301
1800,1.045760,1.780358
2000,1.133329,1.921344
2200,1.218294,2.067844
2400,1.350586,2.805700
2600,1.480698,3.041927
2800,1.599935,3.149908
3000,1.710847,3.263468
3200,1.812947,3.337707
3400,1.914317,3.536128
3600,2.015887,3.742468
3800,2.113941,3.878776
4000,2.205534,3.945651
4200,2.295827,4.101401
4400,2.384309,4.242302
4600,2.463781,4.212057
4800,2.581943,5.299518
5000,2.724604,6.148329
```

## Example: `concurrent`

```Shell
for N in `seq 10 10 100`; do quaperf -mode concurrent -goroutines $N; done
```

```
2022/11/06 23:10:21 generated 2000 patterns and 10000 events in 775.069298ms
goroutines,patterns,copySecs,events,msPerEvent,elapsedSecs,matched
2022/11/06 23:10:23 added 2000 patterns in 1.891049237s (0.945525 ms per pattern)
10,2000,0.000001,10000,0.001530,0.015304,1000
10,2000,0.000000,10000,0.001531,0.015312,1000
10,2000,0.000002,10000,0.001073,0.010731,1000
10,2000,0.000000,10000,0.001481,0.014807,1000
10,2000,0.000002,10000,0.001642,0.016418,1000
10,2000,0.000000,10000,0.002171,0.021710,1000
10,2000,0.000001,10000,0.005059,0.050588,1000
10,2000,0.000000,10000,0.005227,0.052273,1000
10,2000,0.000001,10000,0.005236,0.052359,1000
10,2000,0.000000,10000,0.005804,0.058039,1000
2022/11/06 23:10:23 generated 2000 patterns and 10000 events in 775.788116ms
goroutines,patterns,copySecs,events,msPerEvent,elapsedSecs,matched
2022/11/06 23:10:25 added 2000 patterns in 1.863584394s (0.931792 ms per pattern)
20,2000,0.000000,10000,0.000966,0.009658,1000
20,2000,0.000000,10000,0.000367,0.003667,1000
20,2000,0.000001,10000,0.001548,0.015475,1000
20,2000,0.000000,10000,0.003567,0.035668,1000
20,2000,0.000000,10000,0.003714,0.037142,1000
20,2000,0.000000,10000,0.000295,0.002951,1000
20,2000,0.000000,10000,0.000986,0.009857,1000
20,2000,0.000001,10000,0.001218,0.012178,1000
...
```

## Example: `showgen`

```Shell
quaperf -mode showgen -min-props 2 | shuf | head
```
```
2022/11/06 23:15:01 generated 2000 patterns and 10000 events in 226.654433ms
event	{"IVGyuNBuIAeY":"oiDNYeHTFgabRMo","UHx":{"GFJJrGJFaiOLxiAolB":[{"dNPtyRw":"BnJUihNILhNWXWbGCSAYTqxwG","jxXSTDdRkzD":"jAUWJmeZmmwzJlqWlIjGjes","lrP":"LAIEPlAkBfanF","nlebmDPLChaSyqwvG":"VOtVcekPHjGlTnBUIcfY"},"UMZSPRYnCNppvdBlLxt","gyEhNCTdsqYVRPhMGimPpfOr","RCEJvBhoNtqkVPDccMiLdVvlCUzf"],"RNBykSBYFWaCWgnN":"JYpZPBnnTfWPhvih","kcyDuZYB":{"WSE":"GKoPOPamJoUfyZPAAgQjW","gugierjupBJiEqlB":"ZLOdstIvOJLrsypn","iuri":"wRHfthPRYBnBXK"},"miIvLadxcUrDZzGE":"MczVaxCGCueyFQuAQo"},"ebknnGfWYfOZD":{"QzIyYqtXxZhaoHlSOV":{"NbWgd":"WhYCraGZkUCjDJAzvqz","SMVfxHhCykSMb":["DHwveLkWKbLiSCkBUuChwILXT"],"SSDJybRgyiYgzmLSQq":"HnDVltclBsdcWLjzDGEjlaa","fqwuSpokM":[-19,-88,"nlBwfKNfNmCnSMKIUhtCO"]},"YXOGjz":["oHMSqFqkInamBYLF","SAIgqdeTXGmiPNKsaB","TAzchBojzyIbLRIPpimnZrG"],"epNxPCvsBUdXnlWaWFL":["uOsbPBUJoPhhjiSHjpYL","MljNUmGUzQqMmnUpFNqm",23,-33],"pFvJOUwIN":-83},"yEWKSFa":{"NmPugxxvGBZpidWPTF":"kIXSrALf","bRZaOzLrUYdTHwwsbk":-48,"nrGHDTxFxAorkHkbu":{"DjfywN":"wckwMHDvMhmnKNvNkkhSloX","XKOzaeujjqiA":"mZPPTMeLgFzhNispfgCjt"}}}
event	{"RBpsOvdtcFAudaMOxq":{"VnMVziwWhRW":"oshkXiNRKM","dommnIybLvio":"ekOAdfUeFiXCIqqm"},"YGNmsoeBYItXgeVWpeV":["PuEmdusSmXynLdFYrTXyMDeEdnTrr","iRiusZwgbZzThKoJvyRvQPJUTI",-43],"qOihRcao":[46,"ezztMSDiMSCAUVZ"],"skaXQXuZJkSPwk":-90}
event	{"EPTz":-71,"HaleWrqI":{"OKQ":"wmyTA","eupVHLWRShsfsjsAN":"ljlGEStQZbKFuMKPUO"},"qsLpuO":{"FQwy":"vJIaEgVEHYilGrdNDu","dKFUGzIalIUN":"amPiJaPqTgM","fWeDGtrAkFG":"IkbibBbawlvPHhxqVgmouCNR","wbsUNVUxwlog":"fwnNjIxURNZvQfKEJiwyVZ"}}
event	{"AjHTBTrbGHCB":"jaOIL","JgcuxANnhdBwtYs":{"XECUSi":"XIKKRAjdh","pSjzOAm":"tIyhcywkyzFkVnBIaOZgziRGqJ"},"LOodVS":{"WBbkUkHKh":"WsAJrNUDGpUAf","YdSizS":["JBIdIcxxZxpbVyHlB","dmtgqH","qwAkPIumxBgzrErRB"]},"WcVYGwr":"WSKFXNCHRMgVDAtvprEwAhlt"}
event	{"RZSTalBTOaGmrSU":{"CBwJRBEXWTPmUtdagod":"SRezln","toTm":"rqymPUAKcZA"},"kWDdX":"VangvirnyGXSaGJ"}
pattern	{"PDBciFz":["NysYjCKkQeQAahikHPguMg"],"RWDdumIQkXMyhmfU":[-80],"nnLD":["aaAIQDEOcalS"],"uHOQadZKLTPRnJArKv":["uSHlzhJbaNhJCo"]}
pattern	{"ySJeAaGyajCQkOPbD":["SkgVlyhj"]}
event	{"esXbqm":"ESUncWzXaHXNAFeSWBAtamEsRo","yImDeEXdDheQjhzzBY":"CxnLHVxGs"}
event	{"QnsHNkeaTUGzXkxwfN":{"Wpj":"nCRCeUfYlwYIeWERQryhrqxxok","yoyRIxNRLlJ":["ulUzZoqXHEVTxbUtTcTbI","GqDrZXN"]},"TuF":"GsiAqTgeIcooWXC"}
event	{"cqh":"rzFIMABXSDNQhZRHBSceZyIAod","hQaC":["XhVNDalSJSLbYeNICDrlVgkHPbhN","uagHt"],"kqmbSQryAKIjrC":"vQOnSxsnlDCnxizxxqJU","zcrTLX":["PmvgSedKpH"]}
```

## To do

- [ ] Characterize the distribution of events and patterns complexity


