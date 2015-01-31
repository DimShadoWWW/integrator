function CreateButtons(series){
    var output="<table class=\"table table-striped\"><thead><tr><th>ID</th><th>RepoTags</th><th>Created</th><th>Actions</th></tr></thead><tbody>";
    for (var i in series.Images.Images) {
        buttons = "<button type=\"button\" href=\"/api/images/del/"+series.Images.Images[i].Id +"\" class=\"btn btn-primary btn-stop-cont\">Remove</button>"
        output+="<tr><td>"+series.Images.Images[i].Id.substring(0,12) +"</td><td>" + series.Images.Images[i].RepoTags[0] +"</td><td class=\"timeago\">" + series.Images.Images[i].Created +"</td><td>"+buttons+"</td></tr>";
    }
    output+="</tbody></table>";
    $('#content').html(output);
}
