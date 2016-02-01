(function() {
  var data, loaderJson, searchWord;

  data = {};

  loaderJson = function() {
    return $.getJSON('.data.json', function(d) {
      return data = d;
    });
  };

  searchWord = function(data) {
    var count, myExp, output, searchTerm;
    searchTerm = $('#search').val();
    myExp = new RegExp(searchTerm, 'igm');
    output = '<ul id=\'result\'>';
    count = 0;
    $.each(data, function(key, val) {
      var i;
      for (i in val.Words) {
        if (val.Words[i].search(myExp) !== -1) {
          output += '<li>';
          output += '<a href="' + val.Path + '">' + val.Path.replace(/\//gi, " - ").replace(/\.html/gi, "") + '</a>';
          output += '</li>';
          count++;
          return;
        }
      }
    });
    output += '</ul>';
    $('#update').html(output);
    $('#count').html(count + " résultat(s)");
  };

  loaderJson();

  $('#search').keyup(function() {
    searchWord(data);
  });

}).call(this);
