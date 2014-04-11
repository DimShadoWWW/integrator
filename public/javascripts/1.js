
var options = {
    lines: {
        show: true
    },
    points: {
        show: true
    },
    xaxis: {
        tickDecimals: 0,
        tickSize: 1
    }
};
var data = [];
$.plot("#placeholder", data, options);
// Fetch one series, adding to what we already have
var alreadyFetched = {};

$("button.fetchSeries").click(function () {
    var button = $(this);
    // Find the URL in the link right next to us, then fetch the data
    var dataurl = button.siblings("a").attr("href");
    function onDataReceived(series) {
        // Extract the first coordinate pair; jQuery has parsed it, so
        // the data is now just an ordinary JavaScript object
        var firstcoordinate = "(" + series.data[0][0] + ", " + series.data[0][1] + ")";
        button.siblings("span").text("Fetched " + series.label + ", first point: " + firstcoordinate);
        // Push the new data onto our existing data array
        if (!alreadyFetched[series.label]) {
            alreadyFetched[series.label] = true;
            data.push(series);
        }
        $.plot("#placeholder", data, options);
    }
    $.ajax({
        url: dataurl,
        type: "GET",
        dataType: "json",
        success: onDataReceived
    });
});
// Initiate a recurring data update
$("button.dataUpdate").click(function () {
    data = [];
    alreadyFetched = {};
    $.plot("#placeholder", data, options);
    var iteration = 0;
    function fetchData() {
        ++iteration;
        function onDataReceived(series) {
            // Load all the data in one pass; if we only got partial
            // data we could merge it with what we already have.
            data = [ series ];
            $.plot("#placeholder", data, options);
        }
        // Normally we call the same URL - a script connected to a
        // database - but in this case we only have static example
        // files, so we need to modify the URL.
        $.ajax({
            url: "data-eu-gdp-growth-" + iteration + ".json",
            type: "GET",
            dataType: "json",
            success: onDataReceived
        });
        if (iteration < 5) {
            setTimeout(fetchData, 1000);
        } else {
            data = [];
            alreadyFetched = {};
        }
    }
    setTimeout(fetchData, 1000);
});
// Load the first series by default, so we don't have an empty plot
$("button.fetchSeries:first").click();
// Add the Flot version string to the footer
$("#footer").prepend("Flot " + $.plot.version + " &ndash; ");

