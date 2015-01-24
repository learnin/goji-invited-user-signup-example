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
    isRelease = !!gutil.env.release;

var paths = {
  js: [
    './src/app.js',
    './bower_components/angular-bootstrap/ui-bootstrap-tpls.js'
  ],
  css: [
    './bower_components/bootstrap/dist/css/bootstrap.min.css',
    './assets/stylesheets/app.css'
  ]
};

watchify.args.fullPaths = false;

var bundler = browserify({
  entries: paths.js,
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

gulp.task('css', function() {
  gulp.src(paths.css)
    .pipe(concat('bundle.css'))
    .pipe(gulpif(isRelease, minifyCSS()))
    .pipe(gulp.dest('./assets/stylesheets'));
});

gulp.task('default', ['browserify', 'css']);
