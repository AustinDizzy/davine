$(function() {
  Highcharts.SparkLine = function(options, callback) {
    var defaultOptions = {
      chart: {
        renderTo: (options.chart && options.chart.renderTo) || this,
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
      val = $('.formatInt')[i].innerText;
      str = numeral(parseInt(val));
      $('.formatInt')[i].innerText = str.format($(this).data("format"));
  });
  $('.jumbotron').geopattern("Davine - Vine Analytics", {
    'color': "#00bf8f",
    'generator': 'squares'
  });
  $('[data-toggle="tooltip"]').tooltip({
    delay: {
      "show": 0,
      "hide": 800
    },
    html: true
  });
  $('.panel-heading').each(function(){
      $(this).parent().find(".panel-header a, .fa-vine").css({
          "color": "#" + $(this).data("color")
      });
      $(this).geopattern($(this).data("name"), {
          color: "#" + $(this).data("color")
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

  var bgData = [];
  while (bgData.length != 2) {
    var arr = []
    while (arr.length < 7) {
      var randomnumber = Math.ceil(Math.random() * 15), found = false;
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
  $('#graphbg-cont').highcharts({
    chart: {
      animation: false,
      type: 'areaspline',
      backgroundColor: '#00bf8f',
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
      margin: 0,
      spacing: [0, 0, 0, 0]
    },
    exporting: {
      enabled: false
    },
    colors: ['#00E7AD', '#168E70'],
    title: {
      text: false
    },
    legend: {
      enabled: false
    },
    xAxis: {
      labels: {
        enabled: false
      },
      tickLength: 0
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
  })
  var data = $("#graphbg-cont").highcharts().getSVG({
    exporting: {
      sourceWidth: $('#report-block .well').get(0).clientWidth,
      sourceHeight: $('#report-block .well').get(0).clientHeight
    }
  });
  $("#graphbg-cont").remove();
  var DOMURL = window.URL || window.webkitURL || window;

  var svg = new Blob([data], {
    type: 'image/svg+xml;charset=utf-8'
  });
  var url = DOMURL.createObjectURL(svg);

  $('#report-block .well').css({
    'background-image': 'url(' + url + ')'
  });
});
