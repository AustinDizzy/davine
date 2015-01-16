$(function(){
    
    $("[data-toggle='tooltip']").tooltip();
    $('ul.nav > li > a[href="'+ window.location +'"]').parent().addClass('active');

    $('ul.nav > li > a').filter(function() {
        return this.href == window.location;
    }).parent().addClass('active');

    $('a[data-vanity][data-user]').each(function(){
        var url = ($(this).attr('data-vanity') === '' ? $(this).attr('data-user') : $(this).attr('data-vanity'));
        $(this).attr('href', '/u/' + url);
    });

    $('.search-show').click(function(){
        $(this).hide();
        $('#search-cont').show();
    });

    $('.search-hide').click(function(){
        $('.search-show').show();
        $('#search-cont').hide();
    });

    $('.formatInt').each(function(i) {
        val = $('.formatInt')[i].innerText;
        str = numeral(parseInt(val));
        $('.formatInt')[i].innerText = str.format($(this).data("format"));
    });

    var menu = $('#main-menu');
    var origOffsetY = menu.offset().top;

    function scroll() {
        if ($(window).scrollTop() >= origOffsetY) {
            $('#main-menu').addClass('navbar-fixed-top');
            $('.main').css({"padding-top": "72px"});
            $('.navbar-brand').fadeIn('fast');
        } else {
            $('#main-menu').removeClass('navbar-fixed-top');
            $('.main').removeAttr('style');
            $('.navbar-brand').fadeOut('fast');
        }
    }

    document.onscroll = scroll;
});

(function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){
    (i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),
    m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)
})(window,document,'script','//www.google-analytics.com/analytics.js','ga');

ga('create', 'UA-58119240-1', 'auto');
ga('require', 'displayfeatures');
ga('send', 'pageview');