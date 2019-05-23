/**
 * Created by cfwr6466 on 15/09/2016.
 */

remark.macros.scale = function (percentage) {
    var url = this;
    return '<img src="' + url + '" style="width: ' + percentage + '" />';
    // Usage:
    //   ![:scale 50%](image.jpg)
    // Outputs:
    //   <img src="image.jpg" style="width: 50%" />
};