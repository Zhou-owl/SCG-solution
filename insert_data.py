import psycopg2
import random
from faker import Faker

fake_th = Faker('th_TH')  # Create an instance of Faker with the Thailand locale
fake_en = Faker('en_US')  # Create an instance of Faker with the English locale

jdbc_url = 'jdbc:postgresql://scg-chr.cjwqh3bcqjqu.ap-southeast-1.rds.amazonaws.com:5432/'
username = 'master'
password = 'password'

# Connect to the database
connection = psycopg2.connect(
    host="scg-chr.cjwqh3bcqjqu.ap-southeast-1.rds.amazonaws.com",
    port="5432",
    user=username,
    password=password,
    dbname="postgres"
)

# Create the "employee" table
cursor = connection.cursor()
cursor.execute('''
CREATE TABLE IF NOT EXISTS employee (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    employee_id VARCHAR(20) NOT NULL,
    email VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    photo_url VARCHAR(200) NOT NULL,
    hire_date DATE NOT NULL,
    department VARCHAR(50) NOT NULL,
    job_title VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    working_location VARCHAR(50) NOT NULL
);
''')
connection.commit()

# Generate and insert dummy data
for _ in range(50):
    first_name = fake_en.first_name()  # English first name
    last_name = fake_en.last_name()  # English last name
    employee_id = f"{random.randint(10000, 99999)}"
    email = fake_en.email()
    phone_number = fake_th.phone_number()[:20]
    photo_url = f"https://example.com/photos/{employee_id}.jpg"
    hire_date = fake_th.date_between(start_date="-10y", end_date="today")
    department = fake_en.job()[:50]  # English job title for department
    job_title = fake_en.job()[:50]  # English job title
    status = random.choice(["active", "inactive"])
    working_location = fake_en.city()  # English city name

    cursor.execute(
        """
        INSERT INTO employee (first_name, last_name, employee_id, email, phone_number, photo_url, hire_date, department, job_title, status, working_location)
        VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
        """,
        (first_name, last_name, employee_id, email, phone_number, photo_url, hire_date, department, job_title, status, working_location)
    )

# Commit the transaction and close the connection
connection.commit()
cursor.close()
connection.close()

print("Dummy data inserted successfully.")
