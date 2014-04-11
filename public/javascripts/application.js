$(document).ready(function(){
    var options = {series: {pie: {show: true, label: {show: true, radius: 1}, offset: {left: -30}}}};

    // Initiate a recurring data update
    var data_containers = [];
    var data_images = [];

    var iteration = 0;
    function fetchData() {
        ++iteration;
        function onDataReceived(series) {
            var output="<table class=\"table table-striped\"><thead><tr><th>ID</th><th>Names</th><th>Image</th><th>Status</th><th>Actions</th></tr></thead><tbody>";
            for (var i in series.Containers.Containers) {
                buttons="";
                if (series.Containers.Containers[i].Status.split(" ")[0] == "Up") {
                    buttons = "<button type=\"button\" onclick=\"$.ajax({url: '/api/containers/stop/"+series.Containers.Containers[i].Id+"',type: 'GET',dataType: 'json'})\" data-loading-text=\"Stopping...\" class=\"btn btn-primary btn-stop-cont\">Stop</button>"
                } else {
                    buttons = "<button type=\"button\" onclick=\"$.ajax({url: '/api/containers/start/"+series.Containers.Containers[i].Id+"',type: 'GET',dataType: 'json'})\" data-loading-text=\"Starting...\" class=\"btn btn-primary btn-start-cont\">Start</button>"
                    buttons += "<button type=\"button\" id=\"del_"+series.Containers.Containers[i].Id+"\" data-loading-text=\"Removing...\" class=\"btn btn-primary btn-del-cont\">Delete</button>"
                }
                output+="<tr><td>"+series.Containers.Containers[i].Id.substring(0,12) +"</td><td>" + series.Containers.Containers[i].Names +"</td><td>" + series.Containers.Containers[i].Image +"</td><td>" + series.Containers.Containers[i].Status +
                "</td><td>"+buttons+"</td></tr>";
            }
            output+="</tbody></table>";
            $('#content').html(output);

            data_containers = [];
            // var legend_containers="<table class=\"table table-striped\"><thead><tr><th>Name</th><th>Value</th></tr></thead><tbody>";
            for (var i in series.Containers.Status) {
                data_containers.push({
                    label:   i,
                    data: series.Containers.Status[i]
                    });
                // legend_containers+="<tr><td class=\"text-muted\">"+ i +"</td><td class=\"text-muted\">" + series.Containers.Status[i] +"</td></tr>";
            }
            // legend_containers+="</tbody></table>";
            // $('#content').html(legend_containers);

            var data_images = [];
            // var legend_images="<table class=\"table table-striped\"><thead><tr><th>Name</th><th>Value</th></tr></thead><tbody>";
            for (var i in series.Images.Status) {
                data_images.push({
                    label:   i,
                    data: series.Images.Status[i]
                    });
                // legend_images+="<tr><td class=\"text-muted\">"+ i +"</td><td class=\"text-muted\">" + series.Images.Status[i] +"</td></tr>";
            }
            // legend_images+="</tbody></table>";
            // $('#content').html(legend_images);

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
    var timerID = setInterval(function(){fetchData()}, 1000);

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