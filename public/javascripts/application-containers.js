function CreateButtons(series){
    var output="<table class=\"table table-striped\"><thead><tr><th>ID</th><th>Names</th><th>Image</th><th>Status</th><th>Actions</th></tr></thead><tbody>";
    for (var i in series.Containers.Containers) {
        buttons="";
        if (series.Containers.Containers[i].Status.split(" ")[0] == "Up") {
            buttons = "<button type=\"button\" onclick=\"$.ajax({url: '/api/containers/stop/"+series.Containers.Containers[i].Id+"',type: 'GET',dataType: 'json'})\" data-loading-text=\"Stopping...\" class=\"btn btn-primary btn-stop-cont\">Stop</button>"
        } else {
            buttons = "<button type=\"button\" onclick=\"$.ajax({url: '/api/containers/start/"+series.Containers.Containers[i].Id+"',type: 'GET',dataType: 'json'})\" data-loading-text=\"Starting...\" class=\"btn btn-primary btn-start-cont\">Start</button>"
            buttons += "<button type=\"button\" onclick=\"$.ajax({url: '/api/containers/del/"+series.Containers.Containers[i].Id+"',type: 'GET',dataType: 'json'})\" data-loading-text=\"Removing...\" class=\"btn btn-primary btn-del-cont\">Delete</button>"
        }
        output+="<tr><td>"+series.Containers.Containers[i].Id.substring(0,12) +"</td><td>" + series.Containers.Containers[i].Names +"</td><td>" + series.Containers.Containers[i].Image +"</td><td>" + series.Containers.Containers[i].Status +
        "</td><td>"+buttons+"</td></tr>";
    }
    output+="</tbody></table>";
    $('#content').html(output);
}