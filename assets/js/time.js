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

// Scheduls list

Date.prototype.formatDDMMYYYY = function() {
    return this.getDate() + "." + (this.getMonth() + 1) + "." + this.getFullYear();
}

function timePretty(inp) {
    while (inp.length < 2) {
        inp = "0" + inp;
    }
    return inp;
}

Date.prototype.formathhmmss = function() {
    return timePretty(this.getHours().toString()) + ":" + timePretty(this.getMinutes().toString()) + ":" + timePretty(this.getSeconds().toString());
}

function loadSchedules() {
    $.getJSON("/api/getschedulings", function (json) {
        var list = $("#infolist");
        var html = "";
        
        json.forEach(function(element) {
            html += '<tr><td>' + element.ID + '</td><td>' + new Date(element.StartDate).formatDDMMYYYY() + ' - ' + new Date(element.EndDate).formatDDMMYYYY() + '</td><td>' + new Date(element.StartTime).formathhmmss() + ' - ' + new Date(element.EndTime).formathhmmss() + '</td><td><button class="btn-floating waves-effect waves-light red">X</button></td></tr>';
        }, this);
        
        list.html(html);
    });
}

loadSchedules();