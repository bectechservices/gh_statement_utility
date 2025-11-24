window.onload = function() {
    new JsDatePick({
        useMode: 2,
        target: "datepicker1",
        dateFormat: "%d-%M-%Y",
        yearsRange: [1900, 2070],
        limitToToday: false,
        cellColorScheme: "beige",
        imgPath: "img/",
        weekStartDay: 1
    });

    new JsDatePick({
        useMode: 2,
        target: "datepicker2",
        dateFormat: "%d-%M-%Y",
        yearsRange: [1900, 2070],
        limitToToday: true,
        cellColorScheme: "beige",
        imgPath: "img/",
        weekStartDay: 1
    });
};