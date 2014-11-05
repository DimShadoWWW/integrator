var iteration = 0;

function CreateButtons(series) {
    var output = "<table class=\"table table-striped\"><thead><tr><th>Names</th><th>Actions</th></tr></thead><tbody>";
    for (var i in series) {
        buttons = "<button type=\"button\" onclick=\"$.ajax({url: '/api/templates/run/" + series[i] + "',type: 'GET',dataType: 'json'})\" data-loading-text=\"Starting...\" class=\"btn btn-primary btn-start-cont\">Start</button>"
        buttons += "<button type=\"button\" onclick=\"EditTemplate('" + series[i] + "')\" class=\"btn btn-primary btn-del-cont\">Edit</button>"
        output += "<tr><td>" + series[i].substring(0, 12) + "</td><td>" + series[i] + "</td><td>" + buttons + "</td></tr>";
    }
    output += "</tbody></table>";
    $('#content').html(output);
}

function fetchData() {
    ++iteration;

    function onDataReceived(series) {
        CreateButtons(series);
    }
    // Normally we call the same URL - a script connected to a
    // database - but in this case we only have static example
    // files, so we need to modify the URL.
    $.ajax({
        url: "/api/templates/list",
        type: "GET",
        dataType: "json",
        success: onDataReceived
    });
}

// $('#editor').hide();
// $('#editor').modal('hide');
fetchData();

/**
 * Extracts a potential form to load from query string
 */
var getRequestedExample = function() {
    var query = window.location.search.substring(1);
    var vars = query.split('&');
    var param = null;
    for (var i = 0; i < vars.length; i++) {
        param = vars[i].split('=');
        if (param[0] === 'example') {
            return param[1];
        }
    }
    return null;
};

/**
 * Displays the form entered by the user
 * (this function runs whenever once per second whenever the user
 * changes the contents of the ACE input field)
 */
var generateForm = function() {
    var values = $('#form').jsonFormValue();

};

function RunTemplate(name) {
    $.ajax({
        url: "/api/templates/run/" + name,
        type: "GET",
        dataType: "json"
    }).done(function(code) {
        formObject["value"] = code;
        formObject["form"].push({
            "type": "submit",
            "title": "Save",
            "onClick": function(evt) {
                evt.preventDefault();
            }
        });

        formObject["onSubmitValid"] = function(values) {
            $.ajax({
                type: "POST",
                url: '/api/templates/save/' + name,
                data: JSON.stringify(values)
            });
            $("#form").html("");
            $('#editor').modal('hide');
        };
        $('#editor').modal('show');
        $('#form').jsonForm(formObject);

    }).fail(function() {
        window.alert('Sorry, I could not contact the server!');
    });
}

function EditTemplate(name) {


    var formObject = {
        schema: {},
        form: []
    };

    function onDataReceived(series) {
        CreateButtons(series);

        data_containers = [];
        for (var i in series.Containers.Status) {
            data_containers.push({
                label: i,
                data: series.Containers.Status[i]
            });
        }

        var data_images = [];
        // var legend_images="<table class=\"table table-striped\"><thead><tr><th>Name</th><th>Value</th></tr></thead><tbody>";
        for (var i in series.Images.Status) {
            data_images.push({
                label: i,
                data: series.Images.Status[i]
            });
        }
        $.plot("#placeholder-containers", data_containers, options);
        $.plot("#placeholder-images", data_images, options);
    }


    function loadSchema(schema) {
        // formObject["schema"] = schema["schema"];
        formObject = schema;
        $.ajax({
            url: "/api/templates/read/" + name,
            type: "GET",
            dataType: "json"
        }).done(function(code) {
            formObject["value"] = code;
            formObject["form"].push({
                "type": "submit",
                "title": "Save",
                "onClick": function(evt) {
                    evt.preventDefault();
                }
            });

            formObject["onSubmitValid"] = function(values) {
                $.ajax({
                    type: "POST",
                    url: '/api/templates/save/' + name,
                    data: JSON.stringify(values)
                });
                $("#form").html("");
                $('#editor').modal('hide');
            };
            $('#editor').modal('show');
            $('#form').jsonForm(formObject);

        }).fail(function() {
            window.alert('Sorry, I could not contact the server!');
        });
    }

    // function loadTemplate(temp) {
    //     formObject["form"][0] = temp;
    // }

    $.ajax({
        url: "/schema.json",
        type: "GET",
        dataType: "json",
        success: loadSchema
    });


    // Reset result pane

};

// function NewTemplate() {
//     // $("#newTemplateForm").dialog({
//     //     buttons: [{
//     //         text: "Ok",
//     //         click: function() {
//     //             $(this).dialog("close");
//     //             EditTemplate($("#hidden-input").val());
//     //         }
//     //     }]
//     // });
//     // $("#newTemplateForm").dialog("close");
//     // EditTemplate($("#hidden-input").val());
// };

$(function() {
    $("#newTemplateForm").dialog({
        title: "Template name",
        modal: true,
        position: { my: "center top", at: "center top", of: window },
        buttons: [{
            text: "Ok",
            click: function() {
                $(this).dialog("close");
                EditTemplate($("#hidden-input").val());
            }
        }]
    });
    $("#newTemplateForm").dialog("close");
    $("#new-template").click(function() {
        // $("#hidden-input").val();
        $("#newTemplateForm").dialog("open");
    });
});