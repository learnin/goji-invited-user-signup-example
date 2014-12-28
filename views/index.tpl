<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8">
    <title>ユーザー登録</title>
  </head>
  <body>
    {{ page.Msg }}<br/>
    <form action="authentication" method="post">
      ユーザーID:<input type="text" id="userId" name="userId"><br/>
      新パスワード:<input type="password" id="password" name="password"><br/>
      新パスワード(確認):<input type="password" id="confirmPassword" name="confirmPassword"><br/>
      氏名:姓<input type="text" id="lastName" name="lastName"><br/>
      名<input type="text" id="firstName" name="firstName"><br/>
      メールアドレス:<input type="text" id="mail" name="mail"><br/>
      <input type="hidden" id="hashKey" name="hashKey" value="{{ page.HashKey }}">
      <input type="submit" value="次へ">
    </form>
  </body>
</html>