# Design

1. aku akan punya node dan juga gate
2. gate bertanggung jawab atas nodes
3. saat gate pertama kali join,ia akan broadcast pesan ke semua node di cluster
   bahwa dia adalah gate,dengan pesan json :
   '''
   {
      "name" : "gate name",
      "address" : "gate address"
   }
   '''
4. setiap pesan yang akan dikirimkan ke node oleh gate akan mempunyai structure :
   '''
   create db : {
      "id" : "untuk identifikasi pesan karena setiap pesan unik",
      "name" : "nama database"
      "migration" : ""
   }

   get db : {
      "id" : ""
      "name" : ""
   }

   delete db : {
      "id" : ""
      "name" : ""
   }

   query db : {
      "id" : "",
      "sql query" : ""
   }

   exec db : {
      "id" : "",
      "sql exec" : ""
   }
