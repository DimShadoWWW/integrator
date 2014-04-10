function RenderAll() {
    $.getJSON('/api/status', function(json) {
        var output="<table class=\"table table-striped\"><thead><tr><th>ID</th><th>Names</th><th>Image</th><th>Status</th><th>Actions</th></tr></thead><tbody>";
        for (var i in json.Containers.Containers) {
            buttons="";
            if (json.Containers.Containers[i].Status.split(" ")[0] == "Up") {
                buttons = "<button type=\"button\" dockid=\"stop_"+json.Containers.Containers[i].Id+"\" data-loading-text=\"Stopping...\" class=\"btn btn-primary btn-stop-cont\">Stop</button>"
                // buttons = "<a href=\"/api/containers/stop/"+json.Containers.Containers[i].Id+"\">Stop</a>";
            } else {
                buttons = "<button type=\"button\" dockid=\"start_"+json.Containers.Containers[i].Id+"\" data-loading-text=\"Starting...\" class=\"btn btn-primary btn-start-cont\">Start</button>"
                buttons += "<button type=\"button\" dockid=\"del_"+json.Containers.Containers[i].Id+"\" data-loading-text=\"Removing...\" class=\"btn btn-primary btn-del-cont\">Delete</button>"
                // buttons = "<a href=\"/api/containers/start/"+json.Containers.Containers[i].Id+"\">Start</a>";
                // buttons += "<a href=\"/api/containers/del/"+json.Containers.Containers[i].Id+"\">Delete</a>";
            }
            output+="<tr><td>"+json.Containers.Containers[i].Id.substring(0,12) +"</td><td>" + json.Containers.Containers[i].Names +"</td><td>" + json.Containers.Containers[i].Image +"</td><td>" + json.Containers.Containers[i].Status +
            "</td><td>"+buttons+"</td></tr>";
        }
        output+="</tbody></table>";

        document.getElementById("content").innerHTML = output;

        var w = 200, //width
        h = 200, //height
        r = 100, //radius
        color = d3.scale.category20c(); //builtin range of colors
        
        var data_containers = [];
        var legend_containers="<table class=\"table table-striped\"><thead><tr><th>Name</th><th>Value</th></tr></thead><tbody>";
        for (var i in json.Containers.Status) {
            data_containers.push({
                label:   i,
                value: json.Containers.Status[i]
                });
            legend_containers+="<tr><td class=\"text-muted\">"+ i +"</td><td class=\"text-muted\">" + json.Containers.Status[i] +"</td></tr>";
        }
        legend_containers+="</tbody></table>";
        document.getElementById("legend-containers").innerHTML = legend_containers;

        var data_images = [];
        var legend_images="<table class=\"table table-striped\"><thead><tr><th>Name</th><th>Value</th></tr></thead><tbody>";
        for (var i in json.Images.Status) {
            data_images.push({
                label:   i,
                value: json.Images.Status[i]
                });
            legend_images+="<tr><td class=\"text-muted\">"+ i +"</td><td class=\"text-muted\">" + json.Images.Status[i] +"</td></tr>";
        }
        legend_images+="</tbody></table>";
        document.getElementById("legend-images").innerHTML = legend_images;

        // var parent = d3.select("#placeholder-containers").node().parentNode;
        d3.select("#placeholder-containers").node().removeChild(d3.select("#placeholder-containers").node().childNodes[0]);
        var vis_containers = d3.select("#placeholder-containers").insert("svg:svg").data([data_containers]).attr("width", w).attr("height", h).append("svg:g").attr("transform", "translate(" + r + "," + r + ")") //move the center of the pie chart from 0, 0 to radius, radius
         
        var arc_containers = d3.svg.arc().outerRadius(r);
         
        var pie_containers = d3.layout.pie().value(function(d) { return d.value; }); //we must tell it out to access the value of each element in our data array
         
        var arcs_containers = vis_containers.selectAll("g.slice").data(pie_containers).enter().append("svg:g").attr("class", "slice"); //allow us to style things in the slices (like text)
         
        arcs_containers.append("svg:path").attr("fill", function(d, i) { return color(i); } ).attr("d", arc_containers); //this creates the actual SVG path using the associated data (pie) with the arc drawing function
         
        arcs_containers.append("svg:text").attr("transform", function(d) { //set the label's origin to the center of the arc
            //we have to make sure to set these before calling arc.centroid
            d.innerRadius = 0;
            d.outerRadius = r;
            return "translate(" + arc_containers.centroid(d) + ")"; //this gives us a pair of coordinates like [50, 50]
        }).attr("text-anchor", "middle").text(function(d, i) { return data_containers[i].label; }); //get the label from our original data array

        // var parent = d3.select("#placeholder-images").node().parentNode;
        // d3.select("#placeholder-images").remove();
        d3.select("#placeholder-images").node().removeChild(d3.select("#placeholder-images").node().childNodes[0]);
        var vis_images = d3.select("#placeholder-images").insert("svg:svg").data([data_images]).attr("width", w).attr("height", h).append("svg:g").attr("transform", "translate(" + r + "," + r + ")") //move the center of the pie chart from 0, 0 to radius, radius
         
        var arc_images = d3.svg.arc().outerRadius(r);
         
        var pie_images = d3.layout.pie().value(function(d) { return d.value; }); //we must tell it out to access the value of each element in our data array
         
        var arcs_images = vis_images.selectAll("g.slice").data(pie_images).enter().append("svg:g").attr("class", "slice"); //allow us to style things in the slices (like text)
         
        arcs_images.append("svg:path").attr("fill", function(d, i) { return color(i); } ).attr("d", arc_images); //this creates the actual SVG path using the associated data (pie) with the arc drawing function
         
        arcs_images.append("svg:text").attr("transform", function(d) { //set the label's origin to the center of the arc
            //we have to make sure to set these before calling arc.centroid
            d.innerRadius = 0;
            d.outerRadius = r;
            return "translate(" + arc_images.centroid(d) + ")"; //this gives us a pair of coordinates like [50, 50]
        }).attr("text-anchor", "middle").text(function(d, i) { return data_images[i].label; }); //get the label from our original data array
    });
}

$(document).ready(function(){
  $('#btn-clean-all').click(function() {
    $(this).button('loading');
    $.getJSON('/api/clean', function(json) {
        if (json.status == 0) {
            RenderAll();
        }
        $(this).button('reset');
    });
  });  

  $('#btn-clean-cont').click(function() {
    $(this).button('loading');
    $.getJSON('/api/containers/clean', function(json) {
        if (json.status == 0) {
            RenderAll();
        }
        $(this).button('reset');
    });
  });

  $('#btn-clean-images').click(function() {
    $(this).button('loading');
    $.getJSON('/api/images/clean', function(json) {
        if (json.status == 0) {
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
        if (json.status == 0) {
            RenderAll();
        }
    });
  });

  $('.btn-stop-cont').click(function() {
    // $(this).button('stopping');
    var id = $(this).attr('id').split("_")[1];
    $.getJSON('/api/containers/stop/'+id, function(json) {
        if (json.status == 0) {
            RenderAll();
        }
    });
  });

  $('.btn-del-cont').click(function() {
    // $(this).button('deleting');
    var id = $(this).attr('id').split("_")[1];
    $.getJSON('/api/containers/del/'+id, function(json) {
        if (json.status == 0) {
            RenderAll();
        }
    });
  });

  RenderAll();
  var timerID = setInterval(function(){RenderAll()}, 10 * 1000); // 60 * 1000 milsec
});