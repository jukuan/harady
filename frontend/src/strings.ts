export const S = {
  appName: 'Harady',
  tagline: 'Гульня ў гарады',

  // landing
  yourName: 'Тваё імя',
  namePlaceholder: 'Напрыклад, Алесь',
  createRoom: 'Стварыць пакой',
  joinRoom: 'Далучыцца',
  roomCode: 'Код пакоя',
  roomCodePlaceholder: 'ABC123',
  or: 'або',
  createNew: 'Стварыць новы пакой',
  invalidCode: 'Няправільны код пакоя',
  nameRequired: 'Увядзі імя',

  // lobby
  players: 'Гульцы',
  you: 'ты',
  host: 'гаспадар',
  bot: 'бот',
  addBot: 'Дадаць бота',
  start: 'Пачаць',
  waiting: 'Чакаем гаспадара…',
  needPlayers: (n: number) => `Патрэбна хаця б ${n} гульні`,
  copyLink: 'Скапіраваць спасылку',
  copied: 'Скапіравана!',
  shareHint: 'Падзяліся спасылкай з сябрамі',

  // game
  round: 'Раунд',
  actor: 'Актор',
  seconds: 'с',
  yourCity: 'Тваё слова',
  youAreActor: 'Ты — актор!',
  giveClue: 'Дай падказку',
  cluePlaceholder: 'Напішы падказку…',
  guessPlaceholder: 'Напішы свой адказ…',
  guess: 'Адказаць',
  send: 'Даслаць',
  waitingActor: 'Чакаем падказку…',
  correct: 'Правільна!',
  wasCity: (c: string) => `Гэта быў ${c}`,
  nobodyGuessed: 'Ніхто не адгадаў',
  endGame: 'Скончыць гульню',
  youAreGuesser: 'Адгадай горад!',

  // end
  gameOver: 'Гульня скончана',
  scoreboard: 'Вынікі',
  playAgain: 'Гуляць зноў',
  backHome: 'На галоўную',

  // connection
  connecting: 'Злучэнне…',
  reconnecting: 'Перападключэнне…',
  disconnected: 'Сувязь страчана',
  leaveRoom: 'Выйсці',

  // errors
  roomFull: 'Пакой поўны',
  genericError: 'Нешта пайшло не так',
} as const
