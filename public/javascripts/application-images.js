function CreateButtons(series){
    var output="<table class=\"table table-striped\"><thead><tr><th>ID</th><th>Names</th><th>Image</th><th>Status</th><th>Actions</th></tr></thead><tbody>";
    for (var i in series.Images.Images) {
        buttons = "<button type=\"button\" href=\"/buildimage.html\" class=\"btn btn-primary btn-stop-cont\">Stop</button>"
        output+="<tr><td>"+series.Images.Images[i].Id.substring(0,12) +"</td><td>" + series.Images.Images[i].Names +"</td><td>" + series.Images.Images[i].Image +"</td><td>" + series.Images.Images[i].Status +
        "</td><td>"+buttons+"</td></tr>";
    }
    output+="</tbody></table>";
    $('#content').html(output);
}
