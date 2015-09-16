$(function() {
    Highcharts.SparkLine = function(options, callback) {
        var defaultOptions = {
            chart: {
                renderTo: (options.chart && options.chart.renderTo) ||
                    this,
                backgroundColor: null,
                borderWidth: 0,
                type: 'area',
                margin: [2, 0, 2, 0],
                width: 120,
                height: 20,
                style: {
                    overflow: 'visible'
                },
                skipClone: true
            },
            exporting: {
                enabled: false
            },
            title: {
                text: ''
            },
            credits: {
                enabled: false
            },
            xAxis: {
                labels: {
                    enabled: false
                },
                title: {
                    text: null
                },
                startOnTick: false,
                endOnTick: false,
                tickPositions: []
            },
            yAxis: {
                endOnTick: false,
                startOnTick: false,
                labels: {
                    enabled: false
                },
                title: {
                    text: null
                },
                tickPositions: [0]
            },
            legend: {
                enabled: false
            },
            tooltip: {
                enabled: false
            },
            plotOptions: {
                series: {
                    animation: false,
                    lineWidth: 1,
                    shadow: false,
                    states: {
                        hover: {
                            lineWidth: 1
                        }
                    },
                    marker: {
                        radius: 1,
                        states: {
                            hover: {
                                radius: 2
                            }
                        }
                    },
                    fillOpacity: 0.25
                },
                column: {
                    negativeColor: '#910000',
                    borderColor: 'silver'
                }
            }
        };
        options = Highcharts.merge(defaultOptions, options);
        return new Highcharts.Chart(options, callback);
    };
    var $tds = $("span[data-sparkline]"),
        fullLen = $tds.length,
        n = 0;

    function doChunk() {
        var time = +new Date(),
            i,
            len = $tds.length,
            $td,
            stringdata,
            arr,
            data,
            chart;
        for (i = 0; i < len; i += 1) {
            $td = $($tds[i]);
            stringdata = $td.data('sparkline');
            arr = stringdata.split('; ');
            data = $.map(arr[0].split(', '), parseFloat);
            chart = {};
            if (arr[1]) {
                chart.type = arr[1];
            }
            $td.highcharts('SparkLine', {
                series: [{
                    data: data,
                    pointStart: 1
                }],
                chart: chart
            });
            n += 1;
            // If the process takes too much time, run a timeout to allow interaction with the browser
            if (new Date() - time > 500) {
                $tds.splice(0, i + 1);
                setTimeout(doChunk, 0);
                break;
            }
        }
    }
    doChunk();
    $('.formatInt').each(function(i) {
        var elm = $($('.formatInt').get(i));
        var val = elm.text();
        var str = numeral(parseInt(val));
        elm.text(str.format($(this).data("format")));
    });
    $('a[data-vanity][data-user]').each(function() {
        var url = ($(this).attr('data-vanity') === '' ? $(this).attr('data-user') : $(this).attr('data-vanity'));
        $(this).attr('href', '/u/' + url);
    });
    if(!document.location.pathname.startsWith("/u/")) {
        $('#head-hero').geopattern(document.title, {
            'color': "#00bf8f",
            'generator': 'squares'
        });
    }
    $('[data-toggle="tooltip"]').tooltip({
        delay: {
            "show": 0,
            "hide": 800
        },
        html: true
    });
    $('.panel-heading').each(function() {
        var color = $(this).data("color") || "00bf8f";
        if(color.startsWith("0x")) color = color.substring(2)
        $(this).parent().find(".panel-header a, .fa-vine").css({
            "color": "#" + color
        });
        $(this).geopattern($(this).data("name"), {
            color: "#" + color
        });
    });
    if ($(window).width() < 768) {
        $('iframe[src|=https://vine.co/v/]').attr('height', '190');
    }
    $(window).resize(function() {
        if ($(window).width() < 768) {
            $('iframe[src|=https://vine.co/v/]').attr('height', '190');
        }
    });
    $("#user-submit").submit(function(e) {
        e.preventDefault();
        var user = $("input[name='user-link']").val();
        var match = user.match(
            /^(?!\s*$)(?:https?\:\/\/vine.co\/(?:u\/)?)?([a-zA-Z_0-9\.]*$)/
        );
        if (match) {
            $.post("/user?id=" + match[1], function(data) {
                if (data.stored || data.queued) {
                    setTimeout(function(){
                      document.location = "/u/" + match[1];
                    }, 750)
                } else if (!data.exists) {
                    alert("Sorry, the user " + match[1] + " doesn't seem to exist on Vine. Please try again.");
                } else if (!data.stored && data.exists) {
                    alert("Sorry, there was an error adding user " + match[1] + " to our database. If this problem persists, please contact us.");
                }
            });
        }
    });
});

$.fn.chartBG = function(opts){
    return this.each(function(){
        var elm = document.createElement("div"),
            bgData = [],
            defaultTypes = ["area", "areaspline", "line", "column", "spline", "scatter"],
            chartType;
        if(opts == null) opts = {};
        if (Array.isArray(opts.type)) {
            chartType = opts.type[Math.floor(Math.random() * opts.type.length)];
        } else if(typeof opts.type == "string") {
            chartType = opts.type;
        } else if(opts.type == null) {
            chartType = defaultTypes[Math.floor(Math.random() * defaultTypes.length)];
        }

        while (bgData.length != 2) {
            var arr = [], n = (!isNaN(opts.n) ? opts.n : 7);
            while (arr.length < n) {
                var randomnumber = Math.ceil(Math.random() * (n*Math.E)), found = false;
                for (var i = 0; i < arr.length; i++) {
                    if (arr[i] == randomnumber) {
                        found = true;
                        break;
                    }
                }
                if (!found) arr[arr.length] = randomnumber;
            }
            bgData.push(arr);
        }
        $(elm).highcharts({
            chart: {
                animation: false,
                type: chartType,
                backgroundColor: opts.bgColor || '#00bf8f',
                events: {
                    load: function() {
                        this.yAxis[0].update({
                            labels: {
                                enabled: false
                            },
                            title: {
                                text: null
                            }
                        });
                    }
                },
                margin: [0,0,0,0],
                spacing: [0, 0, 0, 0],
                plotBorderWidth: 0
            },
            exporting: {
                enabled: false
            },
            colors: opts.colors || ['#00E7AD', '#168E70'],
            title: {
                text: false
            },
            legend: {
                enabled: false
            },
            xAxis: {
                type: 'datetime',
                gridLineWidth: 0,
                labels: {
                    enabled: false
                },
                tickLength: 0,
                minPadding: 0,
                maxPadding: 0
            },
            yAxis: {
                lineWidth: 0,
                minorGridLineWidth: 0,
                lineColor: 'transparent',
                gridLineColor: 'transparent',

                labels: {
                    enabled: false
                },
                minorTickLength: 0,
                tickLength: 0
            },
            tooltip: {
                enabled: false
            },

            credits: {
                enabled: false
            },
            plotOptions: {
                areaspline: {
                    enableMouseTracking: false,
                    fillOpacity: 0.5
                }
            },
            series: [{
                data: bgData[0]
            }, {
                data: bgData[1]
            }]
        });

        var data = $(elm).highcharts().getSVG({
            exporting: {
                sourceWidth: opts.width || this.clientWidth,
                sourceHeight: opts.height || this.clientHeight
            }
        }),
        DOMURL = window.URL || window.webkitURL || window,
        svg = new Blob([data], {
            type: 'image/svg+xml;charset=utf-8'
        }),
        url = DOMURL.createObjectURL(svg);

        if(this.tagName === 'IMG') {
            $(this).attr('src', url);
        } else {
            $(this).css({
                'background-image': 'url(' + url + ')'
            });
        }
    });
}
