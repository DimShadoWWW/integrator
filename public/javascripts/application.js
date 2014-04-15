$(document).ready(function(){
    $("td.timeago").timeago();
    var options = {series: {
                    pie: {
                        show: true,
                        label: {
                            show: true,
                            radius: 3/4,
                            formatter: function(label, series){
                                return '<div style="font-size:8pt;text-align:center;padding:2px;color:rgba(111, 73, 73, 1);">'+label+'<br/>'+Math.round(series.percent)+'% ('+series.data[0][1]+')</div>';
                            }
                        },
                        background: {
                            opacity: 0.5
                        },
                        offset: {
                            left: -25
                        }
                    }
                }
            };

    // Initiate a recurring data update
    var data_containers = [];
    var data_images = [];

    var iteration = 0;
    function fetchData() {
        ++iteration;
        function onDataReceived(series) {
            CreateButtons(series);

            data_containers = [];
            for (var i in series.Containers.Status) {
                data_containers.push({
                    label:   i,
                    data: series.Containers.Status[i]
                    });
            }

            var data_images = [];
            // var legend_images="<table class=\"table table-striped\"><thead><tr><th>Name</th><th>Value</th></tr></thead><tbody>";
            for (var i in series.Images.Status) {
                data_images.push({
                    label:   i,
                    data: series.Images.Status[i]
                    });
            }
            $.plot("#placeholder-containers", data_containers, options);
            $.plot("#placeholder-images", data_images, options);
        }
        // Normally we call the same URL - a script connected to a
        // database - but in this case we only have static example
        // files, so we need to modify the URL.
        $.ajax({
            url: "/api/status",
            type: "GET",
            dataType: "json",
            success: onDataReceived
        });
    }

    fetchData();
    var timerID = setInterval(function(){fetchData()}, 3000);

    $('#btn-clean-all').click(function() {
        $(this).button('loading');
        $.getJSON('/api/clean', function(json) {
            if (series.status == 0) {
                RenderAll();
            }
            $(this).button('reset');
        });
    });  

    $('#btn-clean-cont').click(function() {
        $(this).button('loading');
        $.getJSON('/api/containers/clean', function(json) {
            if (series.status == 0) {
                RenderAll();
            }
            $(this).button('reset');
        });
    });

    $('#btn-clean-images').click(function() {
        $(this).button('loading');
        $.getJSON('/api/images/clean', function(json) {
            if (series.status == 0) {
                RenderAll();
            }
            $(this).button('reset');
        });
    });  

    $('.btn-start-cont').click(function() {
        // $(this).button('starting');
        var id = $(this).attr('id').split("_")[1];
        alert("Start" + id);
        $.getJSON('/api/containers/start/'+id, function(json) {
            if (series.status == 0) {
                RenderAll();
            }
        });
    });

    $('.btn-stop-cont').click(function() {
        // $(this).button('stopping');
        var id = $(this).attr('id').split("_")[1];
        $.getJSON('/api/containers/stop/'+id, function(json) {
            if (series.status == 0) {
                RenderAll();
            }
        });
    });

    $('.btn-del-cont').click(function() {
        // $(this).button('deleting');
        var id = $(this).attr('id').split("_")[1];
        $.getJSON('/api/containers/del/'+id, function(json) {
            if (series.status == 0) {
                RenderAll();
            }
        });
    });
});