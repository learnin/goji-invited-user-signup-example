<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8">
    <title>ユーザー登録</title>
  </head>
  <body>
    {{ form.Msg|safe }}<br/>
    <form action="execute" method="post">
      ユーザーID:<input type="text" id="userId" name="userId"><br/>
      パスワード:<input type="password" id="password" name="password"><br/>
      パスワード(確認):<input type="password" id="confirmPassword" name="confirmPassword"><br/>
      氏名:姓<input type="text" id="lastName" name="lastName"><br/>
      名<input type="text" id="firstName" name="firstName"><br/>
      メールアドレス:<input type="text" id="mail" name="mail"><br/>
      <input type="hidden" id="hashKey" name="hashKey" value="{{ form.HashKey }}">
      <input type="submit" value="次へ">
    </form>
  </body>
</html>