package store

import "github.com/jukuan/harady/backend/internal/models"

// SeedCities is the starter pack. Belarusian names, single-language deployment.
// Сities are grouped by region/country only for readability; order does not matter.
//
// The list is intentionally generous — 200+ entries — because the game is more
// fun when players can't memorise the answer set.
func SeedCities() []models.City {
	return []models.City{
		// ================= BELARUS — Minsk Region =================
		{Name: "Мінск", Region: "Мінская вобласць"},
		{Name: "Барысаў", Region: "Мінская вобласць"},
		{Name: "Салігорск", Region: "Мінская вобласць"},
		{Name: "Маладзечна", Region: "Мінская вобласць"},
		{Name: "Жодзіна", Region: "Мінская вобласць"},
		{Name: "Слуцк", Region: "Мінская вобласць"},
		{Name: "Дзяржынск", Region: "Мінская вобласць"},
		{Name: "Вілейка", Region: "Мінская вобласць"},
		{Name: "Смалявічы", Region: "Мінская вобласць"},
		{Name: "Мар’іна Горка", Region: "Мінская вобласць"},
		{Name: "Фаніпаль", Region: "Мінская вобласць"},
		{Name: "Стоўбцы", Region: "Мінская вобласць"},
		{Name: "Заслаўе", Region: "Мінская вобласць"},
		{Name: "Нясвіж", Region: "Мінская вобласць"},
		{Name: "Лагойск", Region: "Мінская вобласць"},
		{Name: "Беразіно", Region: "Мінская вобласць"},
		{Name: "Любань", Region: "Мінская вобласць"},
		{Name: "Клецк", Region: "Мінская вобласць"},
		{Name: "Старыя Дарогі", Region: "Мінская вобласць"},
		{Name: "Узда", Region: "Мінская вобласць"},
		{Name: "Чэрвень", Region: "Мінская вобласць"},
		{Name: "Капыль", Region: "Мінская вобласць"},
		{Name: "Валожын", Region: "Мінская вобласць"},
		{Name: "Крупкі", Region: "Мінская вобласць"},
		{Name: "Мядзел", Region: "Мінская вобласць"},
		{Name: "Бяроза", Region: "Мінская вобласць"},
		{Name: "Клецк", Region: "Мінская вобласць"},
		{Name: "Любань", Region: "Мінская вобласць"},
		{Name: "Нясвіж", Region: "Мінская вобласць"},
		{Name: "Салігорск", Region: "Мінская вобласць"},

		// ================= BELARUS — Brest Region =================
		{Name: "Брэст", Region: "Брэсцкая вобласць"},
		{Name: "Баранавічы", Region: "Брэсцкая вобласць"},
		{Name: "Пінск", Region: "Брэсцкая вобласць"},
		{Name: "Кобрын", Region: "Брэсцкая вобласць"},
		{Name: "Бяроза", Region: "Брэсцкая вобласць"},
		{Name: "Лунінец", Region: "Брэсцкая вобласць"},
		{Name: "Івацэвічы", Region: "Брэсцкая вобласць"},
		{Name: "Пружаны", Region: "Брэсцкая вобласць"},
		{Name: "Драгічын", Region: "Брэсцкая вобласць"},
		{Name: "Ганцавічы", Region: "Брэсцкая вобласць"},
		{Name: "Жабінка", Region: "Брэсцкая вобласць"},
		{Name: "Камянец", Region: "Брэсцкая вобласць"},
		{Name: "Маларыта", Region: "Брэсцкая вобласць"},
		{Name: "Столін", Region: "Брэсцкая вобласць"},
		{Name: "Ляхавічы", Region: "Брэсцкая вобласць"},
		{Name: "Баранавічы", Region: "Брэсцкая вобласць"},
		{Name: "Давыд-Гарадок", Region: "Брэсцкая вобласць"},
		{Name: "Высокае", Region: "Брэсцкая вобласць"},
		{Name: "Косава", Region: "Брэсцкая вобласць"},
		{Name: "Тэлеханы", Region: "Брэсцкая вобласць"},

		// ================= BELARUS — Vitebsk Region =================
		{Name: "Віцебск", Region: "Віцебская вобласць"},
		{Name: "Орша", Region: "Віцебская вобласць"},
		{Name: "Наваполацк", Region: "Віцебская вобласць"},
		{Name: "Полацк", Region: "Віцебская вобласць"},
		{Name: "Паставы", Region: "Віцебская вобласць"},
		{Name: "Глыбокае", Region: "Віцебская вобласць"},
		{Name: "Лепель", Region: "Віцебская вобласць"},
		{Name: "Дуброўна", Region: "Віцебская вобласць"},
		{Name: "Браслаў", Region: "Віцебская вобласць"},
		{Name: "Мёры", Region: "Віцебская вобласць"},
		{Name: "Верхнядзвінск", Region: "Віцебская вобласць"},
		{Name: "Докшыцы", Region: "Віцебская вобласць"},
		{Name: "Сянно", Region: "Віцебская вобласць"},
		{Name: "Талачын", Region: "Віцебская вобласць"},
		{Name: "Чашнікі", Region: "Віцебская вобласць"},
		{Name: "Барань", Region: "Віцебская вобласць"},
		{Name: "Полацк", Region: "Віцебская вобласць"},
		{Name: "Наваполацк", Region: "Віцебская вобласць"},
		{Name: "Орша", Region: "Віцебская вобласць"},
		{Name: "Віцебск", Region: "Віцебская вобласць"},

		// ================= BELARUS — Gomel Region =================
		{Name: "Гомель", Region: "Гомельская вобласць"},
		{Name: "Мазыр", Region: "Гомельская вобласць"},
		{Name: "Жлобін", Region: "Гомельская вобласць"},
		{Name: "Рэчыца", Region: "Гомельская вобласць"},
		{Name: "Светлагорск", Region: "Гомельская вобласць"},
		{Name: "Калінкавічы", Region: "Гомельская вобласць"},
		{Name: "Рагачоў", Region: "Гомельская вобласць"},
		{Name: "Добруш", Region: "Гомельская вобласць"},
		{Name: "Хойнікі", Region: "Гомельская вобласць"},
		{Name: "Нараўля", Region: "Гомельская вобласць"},
		{Name: "Ветка", Region: "Гомельская вобласць"},
		{Name: "Чачэрск", Region: "Гомельская вобласць"},
		{Name: "Ельск", Region: "Гомельская вобласць"},
		{Name: "Буда-Кашалёва", Region: "Гомельская вобласць"},
		{Name: "Петрыкаў", Region: "Гомельская вобласць"},
		{Name: "Ле’льчыцы", Region: "Гомельская вобласць"},
		{Name: "Жыткавічы", Region: "Гомельская вобласць"},
		{Name: "Мазыр", Region: "Гомельская вобласць"},
		{Name: "Рэчыца", Region: "Гомельская вобласць"},
		{Name: "Светлагорск", Region: "Гомельская вобласць"},

		// ================= BELARUS — Grodno Region =================
		{Name: "Гродна", Region: "Гродзенская вобласць"},
		{Name: "Ліда", Region: "Гродзенская вобласць"},
		{Name: "Слонім", Region: "Гродзенская вобласць"},
		{Name: "Ваўкавыск", Region: "Гродзенская вобласць"},
		{Name: "Смаргонь", Region: "Гродзенская вобласць"},
		{Name: "Навагрудак", Region: "Гродзенская вобласць"},
		{Name: "Ашмяны", Region: "Гродзенская вобласць"},
		{Name: "Шчучын", Region: "Гродзенская вобласць"},
		{Name: "Масты", Region: "Гродзенская вобласць"},
		{Name: "Астравец", Region: "Гродзенская вобласць"},
		{Name: "Дзятлава", Region: "Гродзенская вобласць"},
		{Name: "Карэлічы", Region: "Гродзенская вобласць"},
		{Name: "Іўе", Region: "Гродзенская вобласць"},
		{Name: "Свіслач", Region: "Гродзенская вобласць"},
		{Name: "Бераставіца", Region: "Гродзенская вобласць"},
		{Name: "Воранава", Region: "Гродзенская вобласць"},
		{Name: "Ліда", Region: "Гродзенская вобласць"},
		{Name: "Слонім", Region: "Гродзенская вобласць"},
		{Name: "Ваўкавыск", Region: "Гродзенская вобласць"},
		{Name: "Навагрудак", Region: "Гродзенская вобласць"},

		// ================= BELARUS — Mogilev Region =================
		{Name: "Магілёў", Region: "Магілёўская вобласць"},
		{Name: "Бабруйск", Region: "Магілёўская вобласць"},
		{Name: "Асіповічы", Region: "Магілёўская вобласць"},
		{Name: "Горкі", Region: "Магілёўская вобласць"},
		{Name: "Крычаў", Region: "Магілёўская вобласць"},
		{Name: "Быхаў", Region: "Магілёўская вобласць"},
		{Name: "Касцюковічы", Region: "Магілёўская вобласць"},
		{Name: "Шклоў", Region: "Магілёўская вобласць"},
		{Name: "Мсціслаў", Region: "Магілёўская вобласць"},
		{Name: "Чавусы", Region: "Магілёўская вобласць"},
		{Name: "Бялынічы", Region: "Магілёўская вобласць"},
		{Name: "Кіраўск", Region: "Магілёўская вобласць"},
		{Name: "Чэрыкаў", Region: "Магілёўская вобласць"},
		{Name: "Слаўгарад", Region: "Магілёўская вобласць"},
		{Name: "Клічаў", Region: "Магілёўская вобласць"},
		{Name: "Круглае", Region: "Магілёўская вобласць"},
		{Name: "Бабруйск", Region: "Магілёўская вобласць"},
		{Name: "Асіповічы", Region: "Магілёўская вобласць"},
		{Name: "Горкі", Region: "Магілёўская вобласць"},

		// ================= EUROPE =================
		// Germany
		{Name: "Берлін", Region: "Германія"},
		{Name: "Гамбург", Region: "Германія"},
		{Name: "Мюнхен", Region: "Германія"},
		{Name: "Кёльн", Region: "Германія"},
		{Name: "Франкфурт", Region: "Германія"},
		// United Kingdom
		{Name: "Лондан", Region: "Вялікабрытанія"},
		{Name: "Бірмінгем", Region: "Вялікабрытанія"},
		{Name: "Лідс", Region: "Вялікабрытанія"},
		{Name: "Манчэстэр", Region: "Вялікабрытанія"},
		{Name: "Эдынбург", Region: "Вялікабрытанія"},
		// France
		{Name: "Парыж", Region: "Францыя"},
		{Name: "Марсель", Region: "Францыя"},
		{Name: "Ліён", Region: "Францыя"},
		{Name: "Тулуза", Region: "Францыя"},
		{Name: "Ніца", Region: "Францыя"},
		// Norway
		{Name: "Осла", Region: "Нарвегія"},
		{Name: "Берген", Region: "Нарвегія"},
		{Name: "Тронхейм", Region: "Нарвегія"},
		// Sweden
		{Name: "Стакгольм", Region: "Швецыя"},
		{Name: "Гётэборг", Region: "Швецыя"},
		{Name: "Мальмё", Region: "Швецыя"},
		// Denmark
		{Name: "Капенгаген", Region: "Данія"},
		{Name: "Орхус", Region: "Данія"},
		// Finland
		{Name: "Хельсінкі", Region: "Фінляндыя"},
		{Name: "Турку", Region: "Фінляндыя"},
		// Portugal
		{Name: "Лісабон", Region: "Партугалія"},
		{Name: "Порту", Region: "Партугалія"},
		// Spain
		{Name: "Мадрыд", Region: "Іспанія"},
		{Name: "Барселона", Region: "Іспанія"},
		{Name: "Валенсія", Region: "Іспанія"},
		{Name: "Севілья", Region: "Іспанія"},
		// Italy
		{Name: "Рым", Region: "Італія"},
		{Name: "Мілан", Region: "Італія"},
		{Name: "Неапаль", Region: "Італія"},
		{Name: "Турын", Region: "Італія"},
		{Name: "Фларэнцыя", Region: "Італія"},
		// Greece
		{Name: "Афіны", Region: "Грэцыя"},
		{Name: "Салонікі", Region: "Грэцыя"},
		// Poland
		{Name: "Варшава", Region: "Польшча"},
		{Name: "Кракаў", Region: "Польшча"},
		{Name: "Лодзь", Region: "Польшча"},
		{Name: "Гданьск", Region: "Польшча"},
		{Name: "Уроцлаў", Region: "Польшча"},
		// Lithuania
		{Name: "Вільнюс", Region: "Літва"},
		{Name: "Коўна", Region: "Літва"},
		// Latvia
		{Name: "Рыга", Region: "Латвія"},
		{Name: "Даўгаўпілс", Region: "Латвія"},
		// Estonia
		{Name: "Талін", Region: "Эстонія"},
		{Name: "Тарту", Region: "Эстонія"},
		// Ukraine
		{Name: "Кіеў", Region: "Украіна"},
		{Name: "Харкаў", Region: "Украіна"},
		{Name: "Адэса", Region: "Украіна"},
		{Name: "Львоў", Region: "Украіна"},
		{Name: "Днепр", Region: "Украіна"},
		// Czech Republic
		{Name: "Прага", Region: "Чэхія"},
		{Name: "Брно", Region: "Чэхія"},
		// Slovakia
		{Name: "Браціслава", Region: "Славакія"},
		{Name: "Кошыцы", Region: "Славакія"},
		// Hungary
		{Name: "Будапешт", Region: "Венгрыя"},
		{Name: "Дэбрэцэн", Region: "Венгрыя"},
		// Romania
		{Name: "Бухарэст", Region: "Румынія"},
		{Name: "Клуж-Напока", Region: "Румынія"},
		// Bulgaria
		{Name: "Сафія", Region: "Балгарыя"},
		{Name: "Плоўдзіў", Region: "Балгарыя"},
		// Serbia
		{Name: "Бялград", Region: "Сербія"},
		{Name: "Нові Сад", Region: "Сербія"},
		// Croatia
		{Name: "Загрэб", Region: "Харватыя"},
		{Name: "Спліт", Region: "Харватыя"},
		// Slovenia
		{Name: "Любляна", Region: "Славенія"},
		{Name: "Марыбор", Region: "Славенія"},
		// Austria
		{Name: "Вена", Region: "Аўстрыя"},
		{Name: "Грац", Region: "Аўстрыя"},
		// Switzerland
		{Name: "Берн", Region: "Швейцарыя"},
		{Name: "Цюрых", Region: "Швейцарыя"},
		{Name: "Жэнева", Region: "Швейцарыя"},
		// Netherlands
		{Name: "Амстэрдам", Region: "Нідэрланды"},
		{Name: "Ротэрдам", Region: "Нідэрланды"},
		// Belgium
		{Name: "Брусель", Region: "Бельгія"},
		{Name: "Антверпен", Region: "Бельгія"},
		// Ireland
		{Name: "Дублін", Region: "Ірландыя"},
		{Name: "Корк", Region: "Ірландыя"},
		// Iceland
		{Name: "Рэйк’явік", Region: "Ісландыя"},
		// Turkey
		{Name: "Стамбул", Region: "Турцыя"},
		{Name: "Анкара", Region: "Турцыя"},
		{Name: "Ізмір", Region: "Турцыя"},

		// ================= ASIA =================
		// Japan
		{Name: "Токіа", Region: "Японія"},
		{Name: "Ёкагама", Region: "Японія"},
		{Name: "Осака", Region: "Японія"},
		// China
		{Name: "Пекін", Region: "Кітай"},
		{Name: "Шанхай", Region: "Кітай"},
		{Name: "Гуанчжоу", Region: "Кітай"},
		// India
		{Name: "Дэлі", Region: "Індыя"},
		{Name: "Мумбаі", Region: "Індыя"},
		{Name: "Бангалор", Region: "Індыя"},
		// Kazakhstan
		{Name: "Астана", Region: "Казахстан"},
		{Name: "Алматы", Region: "Казахстан"},
		// Uzbekistan
		{Name: "Ташкент", Region: "Узбекістан"},
		{Name: "Самарканд", Region: "Узбекістан"},
		// Turkmenistan
		{Name: "Ашхабад", Region: "Туркменістан"},
		// Tajikistan
		{Name: "Душанбэ", Region: "Таджыкістан"},
		// Kyrgyzstan
		{Name: "Бішкек", Region: "Кыргызстан"},
		// Georgia
		{Name: "Тбілісі", Region: "Грузія"},
		{Name: "Батумі", Region: "Грузія"},
		// Armenia
		{Name: "Ерэван", Region: "Арменія"},
		// Azerbaijan
		{Name: "Баку", Region: "Азербайджан"},
		{Name: "Гянджа", Region: "Азербайджан"},
		// South Korea
		{Name: "Сеул", Region: "Паўднёвая Карэя"},
		{Name: "Пусан", Region: "Паўднёвая Карэя"},
		// Thailand
		{Name: "Бангкок", Region: "Тайланд"},
		// Vietnam
		{Name: "Ханой", Region: "В’етнам"},
		{Name: "Хашымін", Region: "В’етнам"},

		// ================= AMERICAS =================
		// USA
		{Name: "Вашынгтон", Region: "ЗША"},
		{Name: "Нью-Ёрк", Region: "ЗША"},
		{Name: "Лос-Анджэлес", Region: "ЗША"},
		{Name: "Чыкага", Region: "ЗША"},
		{Name: "Сан-Францыска", Region: "ЗША"},
		// Canada
		{Name: "Атава", Region: "Канада"},
		{Name: "Манрэаль", Region: "Канада"},
		{Name: "Таронта", Region: "Канада"},
		{Name: "Ванкувер", Region: "Канада"},
		// Mexico
		{Name: "Мехіка", Region: "Мексіка"},
		{Name: "Гвадалахара", Region: "Мексіка"},
		// Brazil
		{Name: "Бразіліа", Region: "Бразілія"},
		{Name: "Сан-Паўлу", Region: "Бразілія"},
		{Name: "Рыа-дэ-Жанейра", Region: "Бразілія"},
		// Argentina
		{Name: "Буэнас-Айрэс", Region: "Аргенціна"},
		{Name: "Кордава", Region: "Аргенціна"},
		// Chile
		{Name: "Сант’яга", Region: "Чылі"},
		// Colombia
		{Name: "Багата", Region: "Калумбія"},
		{Name: "Медельін", Region: "Калумбія"},
		// Peru
		{Name: "Ліма", Region: "Перу"},
		{Name: "Куска", Region: "Перу"},

		// ================= AFRICA =================
		// Egypt
		{Name: "Каір", Region: "Егіпет"},
		{Name: "Александрыя", Region: "Егіпет"},
		// Algeria
		{Name: "Алжыр", Region: "Алжыр"},
		// South Africa
		{Name: "Прэторыя", Region: "ПАР"},
		{Name: "Ёханэсбург", Region: "ПАР"},
		{Name: "Кейптаўн", Region: "ПАР"},
		// Nigeria
		{Name: "Лагос", Region: "Нігерыя"},
		{Name: "Абуджа", Region: "Нігерыя"},
		// Kenya
		{Name: "Найробі", Region: "Кенія"},
		// Morocco
		{Name: "Рабат", Region: "Марока"},
		{Name: "Касабланка", Region: "Марока"},

		// ================= OCEANIA =================
		// Australia
		{Name: "Канбера", Region: "Аўстралія"},
		{Name: "Сіднэй", Region: "Аўстралія"},
		{Name: "Мельбурн", Region: "Аўстралія"},
		{Name: "Брысбен", Region: "Аўстралія"},
		// New Zealand
		{Name: "Велінгтон", Region: "Новая Зеландыя"},
		{Name: "Окленд", Region: "Новая Зеландыя"},
	}
}

// Seed inserts all cities, skipping duplicates. Returns the number inserted.
func Seed(cs *CityStore) (int, error) {
	added := 0
	for _, c := range SeedCities() {
		_, err := cs.Create(&c)
		if err == ErrDuplicate {
			continue
		}
		if err != nil {
			return added, err
		}
		added++
	}
	return added, nil
}
