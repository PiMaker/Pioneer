// Datepicker init

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

// Popup

function closepopup() {
    $("#popup").closeModal();
}

// CommandID/Close button

var commandID = sessionStorage.getItem("pioneer-scheduler-cmd-id");
$("#btnClose").text("Close Scheduler for #" + commandID);

// Scheduls list

Date.prototype.formatDDMMYYYY = function() {
    return timePretty(this.getDate().toString()) + "." + timePretty((this.getMonth() + 1).toString()) + "." + this.getFullYear().toString();
}

function timePretty(inp) {
    while (inp.length < 2) {
        inp = "0" + inp;
    }
    return inp;
}

Date.prototype.formathhmm = function() {
    return timePretty(this.getHours().toString()) + ":" + timePretty(this.getMinutes().toString());
}

function loadSchedules() {
    $.getJSON("/api/getschedulings/" + commandID, function (json) {
        var list = $("#infolist");
        var html = "";
        
        if (json != null) {
            json.forEach(function(element) {
                html += '<tr><td>' + element.ID + '</td><td>' +
                (element.Dynamic ? '<i class="material-icons">trending_up</i>' : '<i class="material-icons">trending_flat</i>') +
                '</td><td>' + new Date(element.StartDate).formatDDMMYYYY() + ' - ' + new Date(element.EndDate).formatDDMMYYYY() +
                '</td><td>' + /[0-9]{2}\:[0-9]{2}/.exec(element.StartTime)[0] + ' - ' + /[0-9]{2}\:[0-9]{2}/.exec(element.EndTime)[0] +
                '</td><td><button class="btn-floating waves-effect waves-light red" onclick="removeScheduling(' + element.ID + ')">X</button></td></tr>';
            }, this);
        }
        
        list.html(html);
    });
}

setTimeout(loadSchedules, 1);

// Scheduling

function schedule() {
    $("body").addClass("loading");
    
    var sd = dateFromPicka.get("select");
    var ed = dateToPicka.get("select");
    var st = $("#timeFrom").val().split(":");
    var et = $("#timeTo").val().split(":");
    
    var sched = {
        StartDate: sd.year + "-" + timePretty((sd.month + 1).toString()) + "-" + timePretty(sd.date.toString()) + "T00:00:00" + timeOffset,
        EndDate: ed.year + "-" + timePretty((ed.month + 1).toString()) + "-" + timePretty(ed.date.toString()) + "T00:00:00" + timeOffset,
        StartTime: "2016-01-01T" + timePretty(parseInt(st[0]).toString()) + ":" + timePretty(parseInt(st[1]).toString()) + ":00" + timeOffset,
        EndTime: "2016-01-01T" + timePretty(parseInt(et[0]).toString()) + ":" + timePretty(parseInt(et[1]).toString()) + ":00" + timeOffset,
        Dynamic: $("#dynamicChk").is(':checked'),
        CommandID: parseInt(commandID)
    };
    
    var json = JSON.stringify(sched);
    
    $.ajax({
        url: '/api/schedule/' + commandID,
        type: "POST",
        data: json,
        contentType: "application/json",
        dataType: "text",
        success: function (response) {
            loadSchedules();
            $("body").removeClass("loading");
            $("#popup-content").text(response);
            $("#popup").openModal();
        },
        error: function (response) {
            $("body").removeClass("loading");
            $("#popup-content").text(response.responseText);
            $("#popup").openModal();
        }
    });
}

// Removing of Schedulings

function removeScheduling(id) {
    $("body").addClass("loading");
    $.ajax({
        url: '/api/cancelscheduling/' + id,
        type: "POST",
        dataType: "text",
        success: function (response) {
            loadSchedules();
            $("body").removeClass("loading");
            $("#popup-content").text(response);
            $("#popup").openModal();
        },
        error: function (response) {
            $("body").removeClass("loading");
            $("#popup-content").text(response.responseText);
            $("#popup").openModal();
        }
    });
}