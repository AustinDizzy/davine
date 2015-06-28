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
  $('a[data-vanity][data-user]').each(function(){
      var url = ($(this).attr('data-vanity') === '' ? $(this).attr('data-user') : $(this).attr('data-vanity'));
      $(this).attr('href', '/u/' + url);
  });
  $('#head-hero').geopattern(document.title, {
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
      var color = ($(this).data("color") === "" ? "00bf8f" : $(this).data("color"));
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
});