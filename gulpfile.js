'use strict';

var gulp = require('gulp'),
    gutil = require('gulp-util'),
    sourcemaps = require('gulp-sourcemaps'),
    source = require('vinyl-source-stream'),
    buffer = require('vinyl-buffer'),
    watchify = require('watchify'),
    browserify = require('browserify'),
    uglify = require('gulp-uglify'),
    minifyCSS = require('gulp-minify-css'),
    gulpif = require('gulp-if'),
    concat = require('gulp-concat'),
    del = require('del'),
    isRelease = !!gutil.env.release;

var paths = {
  js: {
    browserify: [
      './src/javascripts/app.js',
      './bower_components/angular-bootstrap/ui-bootstrap-tpls.js'
    ],
    notBrowserify: [
      './bower_components/es5-shim/es5-shim.min.js',
      './bower_components/json3/lib/json3.min.js'
    ]
  },
  css: [
    './bower_components/bootstrap/dist/css/bootstrap.min.css',
    './src/stylesheets/app.css'
  ]
};

watchify.args.fullPaths = false;

var bundler = browserify({
  entries: paths.js.browserify,
  debug: !isRelease
}, watchify.args);

var bundle = function() {
  return bundler.bundle()
    .on('error', gutil.log.bind(gutil, 'Browserify Error'))
    .pipe(source('bundle.js'))
    .pipe(buffer())
    .pipe(sourcemaps.init({loadMaps: true}))
    .pipe(gulpif(isRelease, uglify({preserveComments:'some'})))
    .pipe(sourcemaps.write('./'))
    .pipe(gulp.dest('./assets/javascripts'));
};

if (!isRelease) {
  bundler = watchify(bundler);
  bundler.on('update', bundle);
}
bundler.transform('brfs');

gulp.task('browserify', bundle);

gulp.task('notBrowserifyJS', function() {
  gulp.src(paths.js.notBrowserify)
    .pipe(gulp.dest('./assets/javascripts'));
});

gulp.task('css', function() {
  gulp.src(paths.css)
    .pipe(concat('bundle.css'))
    .pipe(gulpif(isRelease, minifyCSS()))
    .pipe(gulp.dest('./assets/stylesheets'));
});

gulp.task('clean', function() {
  del(['log/*.log', 'npm-debug.log', 'assets/javascripts/*', 'assets/stylesheets/*'], function (err, paths) {
    console.log('Deleted files/folders:\n', paths.join('\n'));
  });
});

gulp.task('default', ['browserify', 'notBrowserifyJS', 'css']);
