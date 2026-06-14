package com.ehsatn.magik.data

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

interface DataRepository {
  val data: Flow<List<String>>
}

class DefaultDataRepository : DataRepository {
  override val data: Flow<List<String>> = flow { emit(listOf("Android")) }
}
