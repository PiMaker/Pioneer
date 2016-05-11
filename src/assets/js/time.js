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

var timeSlider = document.getElementById('timeSlider');

noUiSlider.create(timeSlider, {
    range: {
        min: 0,
        max: 60*24
    },
    step: 15,
    start: [ 17*60, 22*60+30 ],
    connect: true,
    tooltips: [ false, false ],
    margin: 15,
    format: wNumb({
		decimals: 0
	})
});

function getTime(minutes) {
    minutes = Math.round(minutes);
    mins = minutes % 60;
    hour = (minutes-mins)/60;
    return (hour < 10 ? "0" + hour : hour) + ":" + (mins < 10 ? "0" + mins : mins);
}

var timeValues = [
	document.getElementById('timeStart'),
	document.getElementById('timeEnd')
];

timeSlider.noUiSlider.on('update', function(values, handle) {
    if (handle < 2 && values[handle] !== null) {
	    timeValues[handle].innerHTML = getTime(+values[handle]);
    }
});