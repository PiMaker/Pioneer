var dateFrom = $("#dateFrom");
var dateTo = $("#dateTo");

dateTo.pickadate({
    selectMonths: true,
    min: new Date(),
    clear: '',
    format: 'ddd, dd. mmm yyyy',
    onStart: function() {
        var date = new Date();
        this.set('select', [ date.getFullYear(), date.getMonth(), date.getDate() ], {muted: true});
    },
    onSet: function(context) {
        var newDate = new Date(context.select);
        var fromDate = new Date(dateFromPicka.get('select').pick);
        if (newDate < fromDate) {
            dateFromPicka.set('select', newDate);
        }
    }
});

dateFrom.pickadate({
    selectMonths: true,
    min: new Date(),
    clear: '',
    format: 'ddd, dd. mmm yyyy',
    onStart: function() {
        var date = new Date();
        this.set('select', [ date.getFullYear(), date.getMonth(), date.getDate() ], {muted: true});
    },
    onSet: function(context) {
        var newDate = new Date(context.select);
        dateToPicka.set('select', newDate);
    }
});

var dateFromPicka = dateFrom.pickadate('picker');
var dateToPicka = dateTo.pickadate('picker');

// Check time compatibility

function timeFallback() {
    $(".timepickerjs").timepicker({ 'step': 1, 'timeFormat': 'H:i' });
}

try {
    var input = document.createElement("input");
    input.type = "time";

    if (input.type === "time") {
        // Supported, set time type
        $(".timepickerjs").attr("type", "time");
    } else {
        timeFallback();
    }
} catch(e) {
    timeFallback();
}

function closepopup() {
    $("#popup").closeModal();
}

function schedule() {
    $("#popup").openModal();
}